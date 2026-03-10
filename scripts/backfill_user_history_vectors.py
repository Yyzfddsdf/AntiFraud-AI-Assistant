#!/usr/bin/env python
# -*- coding: utf-8 -*-

import argparse
import json
import sqlite3
import sys
import urllib.error
import urllib.request
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable, List, Sequence, Tuple


def resolve_project_root() -> Path:
    return Path(__file__).resolve().parents[1]


def default_db_path() -> Path:
    return resolve_project_root() / "DB" / "auth_system.db"


def default_config_path() -> Path:
    return resolve_project_root() / "config" / "config.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="将存量 history_cases 回填到 user_history_vectors 向量索引表。"
    )
    parser.add_argument(
        "--db-path",
        default=str(default_db_path()),
        help="业务主库路径，默认 DB/auth_system.db",
    )
    parser.add_argument(
        "--config-path",
        default=str(default_config_path()),
        help="配置文件路径，默认 config/config.json",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=20,
        help="每批次发送到 embeddings 接口的记录数，默认 20",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=0,
        help="最多处理多少条记录；0 表示不限制",
    )
    parser.add_argument(
        "--user-id",
        default="",
        help="只回填指定 user_id 的历史记录",
    )
    parser.add_argument(
        "--record-id",
        default="",
        help="只回填指定 record_id 的历史记录",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="覆盖已存在的 user_history_vectors 记录；默认仅补缺失记录",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="只打印待处理数量，不实际调用 embeddings 或写数据库",
    )
    return parser.parse_args()


def load_embedding_config(config_path: Path) -> Tuple[str, str, str]:
    with config_path.open("r", encoding="utf-8") as file:
        payload = json.load(file)
    embedding_cfg = payload.get("embedding") or {}
    model = str(embedding_cfg.get("model", "")).strip()
    api_key = str(embedding_cfg.get("api_key", "")).strip()
    base_url = str(embedding_cfg.get("base_url", "")).strip()
    if not model:
        raise ValueError("embedding.model 为空")
    if not base_url:
        raise ValueError("embedding.base_url 为空")
    return model, api_key, base_url.rstrip("/")


def ensure_user_history_vector_table(conn: sqlite3.Connection) -> None:
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS user_history_vectors (
            record_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            embedding_vector TEXT NOT NULL,
            embedding_model TEXT NOT NULL,
            embedding_dimension INTEGER NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            PRIMARY KEY (record_id, user_id)
        )
        """
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_user_history_vectors_user_id ON user_history_vectors(user_id)"
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_user_history_vectors_created_at ON user_history_vectors(created_at)"
    )
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_user_history_vectors_updated_at ON user_history_vectors(updated_at)"
    )
    conn.commit()


def build_history_query(overwrite: bool, user_id: str, record_id: str, limit: int) -> Tuple[str, List[object]]:
    conditions: List[str] = ["1 = 1"]
    params: List[object] = []

    trimmed_user_id = user_id.strip()
    if trimmed_user_id:
        conditions.append("h.user_id = ?")
        params.append(trimmed_user_id)

    trimmed_record_id = record_id.strip()
    if trimmed_record_id:
        conditions.append("h.record_id = ?")
        params.append(trimmed_record_id)

    if not overwrite:
        conditions.append(
            "NOT EXISTS (SELECT 1 FROM user_history_vectors v WHERE v.record_id = h.record_id AND v.user_id = h.user_id)"
        )

    where_clause = " AND ".join(conditions)
    query = (
        "SELECT h.record_id, h.user_id, h.title, h.case_summary, h.scam_type, h.created_at "
        "FROM history_cases h "
        f"WHERE {where_clause} "
        "ORDER BY h.created_at DESC"
    )
    if limit > 0:
        query += " LIMIT ?"
        params.append(limit)
    return query, params


def fetch_history_rows(conn: sqlite3.Connection, overwrite: bool, user_id: str, record_id: str, limit: int) -> List[sqlite3.Row]:
    query, params = build_history_query(overwrite, user_id, record_id, limit)
    conn.row_factory = sqlite3.Row
    cursor = conn.execute(query, params)
    return list(cursor.fetchall())


def normalize_history_text(title: str, case_summary: str, scam_type: str) -> str:
    trimmed_title = title.strip()
    trimmed_summary = case_summary.strip()
    trimmed_scam_type = scam_type.strip()
    if not trimmed_title:
        trimmed_title = trimmed_summary

    segments: List[str] = []
    if trimmed_title:
        segments.append(f"标题: {trimmed_title}")
    if trimmed_summary:
        segments.append(f"案件摘要: {trimmed_summary}")
    if trimmed_scam_type:
        segments.append(f"诈骗类型: {trimmed_scam_type}")
    return "\n".join(segments)


def chunked(items: Sequence[sqlite3.Row], batch_size: int) -> Iterable[Sequence[sqlite3.Row]]:
    normalized_batch_size = max(1, batch_size)
    for index in range(0, len(items), normalized_batch_size):
        yield items[index : index + normalized_batch_size]


def request_embeddings(base_url: str, api_key: str, model: str, texts: Sequence[str]) -> Tuple[List[List[float]], str]:
    endpoint = f"{base_url}/embeddings"
    payload = json.dumps(
        {
            "model": model,
            "input": list(texts),
            "encoding_format": "float",
            "truncate": "NONE",
        }
    ).encode("utf-8")
    request = urllib.request.Request(endpoint, data=payload, method="POST")
    request.add_header("Content-Type", "application/json")
    if api_key:
        request.add_header("Authorization", f"Bearer {api_key}")

    try:
        with urllib.request.urlopen(request, timeout=120) as response:
            raw = response.read()
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"embedding 请求失败，status={exc.code}, body={body}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"embedding 请求发送失败: {exc}") from exc

    payload_obj = json.loads(raw.decode("utf-8"))
    data = payload_obj.get("data") or []
    if not data:
        raise RuntimeError("embedding response is empty")

    sorted_items = sorted(data, key=lambda item: int(item.get("index", 0)))
    vectors: List[List[float]] = []
    for item in sorted_items:
        embedding = item.get("embedding") or []
        if not embedding:
            raise RuntimeError("embedding vector is empty")
        vectors.append([sanitize_float(value) for value in embedding])

    response_model = str(payload_obj.get("model", "")).strip() or model
    return vectors, response_model


def sanitize_float(value: object) -> float:
    try:
        number = float(value)
    except (TypeError, ValueError):
        return 0.0
    if number != number or number in (float("inf"), float("-inf")):
        return 0.0
    return number


def upsert_vectors(conn: sqlite3.Connection, rows: Sequence[sqlite3.Row], vectors: Sequence[List[float]], model: str) -> int:
    now_text = datetime.now(timezone.utc).astimezone().strftime("%Y-%m-%d %H:%M:%S")
    payloads = []
    for row, vector in zip(rows, vectors):
        created_at = str(row["created_at"] or "").strip() or now_text
        payloads.append(
            (
                str(row["record_id"]).strip(),
                str(row["user_id"]).strip(),
                json.dumps(vector, ensure_ascii=False),
                model,
                len(vector),
                created_at,
                now_text,
            )
        )

    conn.executemany(
        """
        INSERT OR REPLACE INTO user_history_vectors (
            record_id,
            user_id,
            embedding_vector,
            embedding_model,
            embedding_dimension,
            created_at,
            updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
        """,
        payloads,
    )
    conn.commit()
    return len(payloads)


def main() -> int:
    args = parse_args()
    db_path = Path(args.db_path).expanduser().resolve()
    config_path = Path(args.config_path).expanduser().resolve()

    if not db_path.exists():
        print(f"[backfill] 数据库不存在: {db_path}", file=sys.stderr)
        return 1
    if not config_path.exists():
        print(f"[backfill] 配置文件不存在: {config_path}", file=sys.stderr)
        return 1

    try:
        model, api_key, base_url = load_embedding_config(config_path)
    except Exception as exc:
        print(f"[backfill] 读取 embedding 配置失败: {exc}", file=sys.stderr)
        return 1

    conn = sqlite3.connect(str(db_path))
    try:
        ensure_user_history_vector_table(conn)
        rows = fetch_history_rows(conn, args.overwrite, args.user_id, args.record_id, args.limit)
        prepared_rows: List[sqlite3.Row] = []
        prepared_texts: List[str] = []
        skipped_empty = 0
        for row in rows:
            text = normalize_history_text(
                str(row["title"] or ""),
                str(row["case_summary"] or ""),
                str(row["scam_type"] or ""),
            )
            if not text:
                skipped_empty += 1
                continue
            prepared_rows.append(row)
            prepared_texts.append(text)

        print(
            "[backfill] 待处理记录: {}，空文本跳过: {}，模式: {}".format(
                len(prepared_rows),
                skipped_empty,
                "overwrite" if args.overwrite else "missing-only",
            )
        )
        if args.dry_run or not prepared_rows:
            return 0

        total_written = 0
        for batch_rows in chunked(prepared_rows, args.batch_size):
            batch_texts = [
                normalize_history_text(
                    str(row["title"] or ""),
                    str(row["case_summary"] or ""),
                    str(row["scam_type"] or ""),
                )
                for row in batch_rows
            ]
            vectors, response_model = request_embeddings(base_url, api_key, model, batch_texts)
            if len(vectors) != len(batch_rows):
                raise RuntimeError(
                    f"embedding 返回数量不匹配: expect={len(batch_rows)} got={len(vectors)}"
                )
            written = upsert_vectors(conn, batch_rows, vectors, response_model)
            total_written += written
            print(f"[backfill] 已写入 {total_written}/{len(prepared_rows)}")

        print(f"[backfill] 完成，累计写入 {total_written} 条 user_history_vectors")
        return 0
    except Exception as exc:
        print(f"[backfill] 失败: {exc}", file=sys.stderr)
        return 1
    finally:
        conn.close()


if __name__ == "__main__":
    sys.exit(main())
