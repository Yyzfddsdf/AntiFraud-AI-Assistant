#!/usr/bin/env python
# -*- coding: utf-8 -*-

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any, Dict, List, Optional, Sequence, Set, Tuple

DEFAULT_ADMIN_INVITE_CODE = "Secret_Admin_Invite_Code_2026"


def resolve_project_root() -> Path:
    return Path(__file__).resolve().parents[1]


def default_input_path() -> Path:
    return resolve_project_root() / "scripts" / "data" / "fraud_cases_final.json"


def default_failure_output_path() -> Path:
    return resolve_project_root() / "scripts" / "data" / "fraud_cases_import_failures.json"


def default_target_groups_path() -> Path:
    return resolve_project_root() / "internal" / "platform" / "config" / "target_groups.json"


def default_scam_types_path() -> Path:
    return resolve_project_root() / "internal" / "platform" / "config" / "scam_types.json"  


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="读取 fraud_cases_final.json，并通过管理员 API 批量导入历史案件库。"
    )
    parser.add_argument(
        "--input",
        default=str(default_input_path()),
        help="输入 JSON 文件路径，默认 scripts/data/fraud_cases_final.json",
    )
    parser.add_argument(
        "--base-url",
        default="http://localhost:8081",
        help="服务地址，默认 http://localhost:8081",
    )
    parser.add_argument(
        "--token",
        default=os.environ.get("ANTI_FRAUD_TOKEN", "").strip(),
        help="管理员 JWT；默认读取环境变量 ANTI_FRAUD_TOKEN。若不传，可结合 --phone 自动登录获取。",
    )
    parser.add_argument(
        "--phone",
        default=os.environ.get("ANTI_FRAUD_PHONE", "").strip(),
        help="登录手机号；默认读取环境变量 ANTI_FRAUD_PHONE",
    )
    parser.add_argument(
        "--sms-code",
        default=os.environ.get("ANTI_FRAUD_SMS_CODE", "000000").strip(),
        help="短信验证码；默认读取环境变量 ANTI_FRAUD_SMS_CODE，未设置时默认 000000",
    )
    parser.add_argument(
        "--invite-code",
        default=os.environ.get("ANTI_FRAUD_INVITE_CODE", DEFAULT_ADMIN_INVITE_CODE).strip(),
        help="管理员升级邀请码；默认读取环境变量 ANTI_FRAUD_INVITE_CODE，否则使用项目默认值",
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=120,
        help="单次 HTTP 请求超时时间（秒），默认 120",
    )
    parser.add_argument(
        "--start-index",
        type=int,
        default=1,
        help="从第几条源数据开始处理，1-based，默认 1",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=0,
        help="最多处理多少条源数据；0 表示不限制",
    )
    parser.add_argument(
        "--sleep-ms",
        type=int,
        default=0,
        help="每次成功请求后额外休眠多少毫秒，默认 0",
    )
    parser.add_argument(
        "--skip-existing",
        action="store_true",
        help="导入前先拉取已有预览列表，按 title+target_group+scam_type 跳过已存在记录",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="只做本地规范化、校验和去重，不实际调用导入 API",
    )
    parser.add_argument(
        "--fail-fast",
        action="store_true",
        help="遇到第一条失败后立即停止",
    )
    parser.add_argument(
        "--failure-output",
        default=str(default_failure_output_path()),
        help="失败记录导出路径；传空字符串表示不导出",
    )
    return parser.parse_args()


def normalize_api_base(base_url: str) -> str:
    trimmed = base_url.strip().rstrip("/")
    if not trimmed:
        raise ValueError("base_url 不能为空")
    if trimmed.endswith("/api"):
        return trimmed
    return trimmed + "/api"


def load_options(config_path: Path, key: str) -> Set[str]:
    with config_path.open("r", encoding="utf-8") as file:
        payload = json.load(file)

    raw_values = payload.get(key)
    if not isinstance(raw_values, list):
        raise ValueError(f"{config_path} 中缺少数组字段 {key}")

    options: Set[str] = set()
    for value in raw_values:
        text = normalize_text(value)
        if text:
            options.add(text)
    return options


def normalize_text(value: Any) -> str:
    if value is None:
        return ""
    return str(value).strip()


def normalize_risk_level(value: Any) -> str:
    level = normalize_text(value)
    alias = {
        "高": "高",
        "高风险": "高",
        "high": "高",
        "中": "中",
        "中风险": "中",
        "medium": "中",
        "低": "低",
        "低风险": "低",
        "low": "低",
    }
    return alias.get(level.lower(), alias.get(level, ""))


def normalize_string_list(value: Any) -> List[str]:
    if value is None:
        return []

    if isinstance(value, list):
        raw_items = value
    else:
        text = normalize_text(value)
        raw_items = [text] if text else []

    normalized: List[str] = []
    seen: Set[str] = set()
    for item in raw_items:
        text = normalize_text(item)
        if not text or text in seen:
            continue
        seen.add(text)
        normalized.append(text)
    return normalized


def first_present(case: Dict[str, Any], keys: Sequence[str]) -> Any:
    for key in keys:
        if key in case:
            return case.get(key)
    return None


def normalize_case(case: Dict[str, Any]) -> Dict[str, Any]:
    title = normalize_text(first_present(case, ["title"]))
    case_description = normalize_text(first_present(case, ["case_description"]))
    if title and case_description and len(case_description) < 12 and title not in case_description:
        case_description = f"{title}：{case_description}"

    payload = {
        "title": title,
        "target_group": normalize_text(first_present(case, ["target_group"])),
        "risk_level": normalize_risk_level(first_present(case, ["risk_level"])),
        "scam_type": normalize_text(first_present(case, ["scam_type"])),
        "case_description": case_description,
        "typical_scripts": normalize_string_list(
            first_present(case, ["typical_scripts", "典型话术"])
        ),
        "keywords": normalize_string_list(first_present(case, ["keywords", "关键词"])),
        "violated_law": normalize_text(first_present(case, ["violated_law", "违反法律", "违法依据"])),
        "suggestion": normalize_text(first_present(case, ["suggestion", "建议"])),
    }
    return payload


def validate_case(
    payload: Dict[str, Any], allowed_target_groups: Set[str], allowed_scam_types: Set[str]
) -> Optional[str]:
    required_fields = [
        "title",
        "target_group",
        "risk_level",
        "scam_type",
        "case_description",
    ]
    for field in required_fields:
        if not normalize_text(payload.get(field)):
            return f"{field} 为空"

    if payload["target_group"] not in allowed_target_groups:
        return f"target_group 非法: {payload['target_group']}"
    if payload["scam_type"] not in allowed_scam_types:
        return f"scam_type 非法: {payload['scam_type']}"
    if payload["risk_level"] not in {"高", "中", "低"}:
        return f"risk_level 非法: {payload['risk_level']}"

    description_length = len(payload["case_description"])
    if description_length < 12:
        return "case_description 过短，少于 12 个字符"
    if description_length > 400:
        return "case_description 过长，超过 400 个字符"
    return None


def load_source_cases(input_path: Path) -> List[Any]:
    with input_path.open("r", encoding="utf-8") as file:
        payload = json.load(file)
    if not isinstance(payload, list):
        raise ValueError("输入 JSON 顶层必须是数组")
    return payload


def build_payload_signature(payload: Dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=False, sort_keys=True)


def build_remote_key(payload: Dict[str, Any]) -> Tuple[str, str, str]:
    return (
        normalize_text(payload.get("title")),
        normalize_text(payload.get("target_group")),
        normalize_text(payload.get("scam_type")),
    )


def request_json(
    url: str,
    method: str,
    token: str,
    timeout: int,
    payload: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    body = None
    if payload is not None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")

    request = urllib.request.Request(url, data=body, method=method)
    request.add_header("Accept", "application/json")
    if body is not None:
        request.add_header("Content-Type", "application/json")
    if token:
        request.add_header("Authorization", f"Bearer {token}")

    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            raw = response.read()
    except urllib.error.HTTPError as exc:
        raw_body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"HTTP {exc.code}: {raw_body}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"请求失败: {exc}") from exc

    if not raw:
        return {}
    try:
        return json.loads(raw.decode("utf-8"))
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"响应不是合法 JSON: {raw[:200]!r}") from exc


def send_sms_code(api_base: str, phone: str, timeout: int) -> None:
    request_json(
        f"{api_base}/auth/sms-code",
        method="POST",
        token="",
        timeout=timeout,
        payload={"phone": phone},
    )


def login_with_sms(api_base: str, phone: str, sms_code: str, timeout: int) -> Tuple[str, Dict[str, Any]]:
    response = request_json(
        f"{api_base}/auth/login",
        method="POST",
        token="",
        timeout=timeout,
        payload={
            "phone": phone,
            "smsCode": sms_code,
        },
    )
    token = normalize_text(response.get("token")) if isinstance(response, dict) else ""
    if not token:
        raise RuntimeError("短信登录成功响应中缺少 token")
    user = response.get("user")
    if not isinstance(user, dict):
        user = {}
    return token, user


def upgrade_to_admin(api_base: str, token: str, invite_code: str, timeout: int) -> None:
    request_json(
        f"{api_base}/upgrade",
        method="POST",
        token=token,
        timeout=timeout,
        payload={"invite_code": invite_code},
    )


def resolve_token(api_base: str, args: argparse.Namespace) -> str:
    explicit_token = normalize_text(args.token)
    if explicit_token:
        return explicit_token

    phone = normalize_text(args.phone)
    sms_code = normalize_text(args.sms_code)
    invite_code = normalize_text(args.invite_code)

    if not phone:
        raise ValueError("未提供 token 时，必须传 --phone 或环境变量 ANTI_FRAUD_PHONE")
    if not sms_code:
        raise ValueError("短信验证码为空")

    send_sms_code(api_base, phone, args.timeout)
    token, user = login_with_sms(api_base, phone, sms_code, args.timeout)
    current_role = normalize_text(user.get("role"))
    print(f"[import] 短信登录成功: phone={phone} role={current_role or '-'}")

    if current_role != "admin":
        if not invite_code:
            raise ValueError("当前账号不是管理员，且未提供邀请码，无法自动升级")
        upgrade_to_admin(api_base, token, invite_code, args.timeout)
        print("[import] 账号已升级为管理员")

    return token


def fetch_existing_remote_keys(api_base: str, token: str, timeout: int) -> Set[Tuple[str, str, str]]:
    response = request_json(
        f"{api_base}/scam/case-library/cases",
        method="GET",
        token=token,
        timeout=timeout,
    )
    items = response.get("cases")
    if not isinstance(items, list):
        raise RuntimeError("历史案件预览接口返回异常，缺少 cases 数组")

    keys: Set[Tuple[str, str, str]] = set()
    for item in items:
        if not isinstance(item, dict):
            continue
        keys.add(
            (
                normalize_text(item.get("title")),
                normalize_text(item.get("target_group")),
                normalize_text(item.get("scam_type")),
            )
        )
    return keys


def write_failures(output_path: Path, failures: Sequence[Dict[str, Any]]) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as file:
        json.dump(list(failures), file, ensure_ascii=False, indent=2)


def prepare_cases(
    raw_cases: Sequence[Any],
    start_index: int,
    limit: int,
    allowed_target_groups: Set[str],
    allowed_scam_types: Set[str],
) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]], int]:
    normalized_cases: List[Dict[str, Any]] = []
    failures: List[Dict[str, Any]] = []
    duplicate_input_count = 0
    seen_signatures: Set[str] = set()

    normalized_start = max(1, start_index)
    sliced_cases = list(raw_cases[normalized_start - 1 :])
    if limit > 0:
        sliced_cases = sliced_cases[:limit]

    for offset, raw_case in enumerate(sliced_cases, start=normalized_start):
        if not isinstance(raw_case, dict):
            failures.append(
                {
                    "source_index": offset,
                    "reason": "源数据不是对象",
                    "source_case": raw_case,
                }
            )
            continue

        payload = normalize_case(raw_case)
        reason = validate_case(payload, allowed_target_groups, allowed_scam_types)
        if reason:
            failures.append(
                {
                    "source_index": offset,
                    "title": payload.get("title", ""),
                    "reason": reason,
                    "normalized_payload": payload,
                    "source_case": raw_case,
                }
            )
            continue

        signature = build_payload_signature(payload)
        if signature in seen_signatures:
            duplicate_input_count += 1
            continue
        seen_signatures.add(signature)

        normalized_cases.append(
            {
                "source_index": offset,
                "payload": payload,
            }
        )

    return normalized_cases, failures, duplicate_input_count


def main() -> int:
    args = parse_args()

    input_path = Path(args.input).expanduser().resolve()
    failure_output = normalize_text(args.failure_output)
    failure_output_path = Path(failure_output).expanduser().resolve() if failure_output else None

    if not input_path.exists():
        print(f"[import] 输入文件不存在: {input_path}", file=sys.stderr)
        return 1

    try:
        api_base = normalize_api_base(args.base_url)
        allowed_target_groups = load_options(default_target_groups_path(), "target_groups")
        allowed_scam_types = load_options(default_scam_types_path(), "scam_types")
        raw_cases = load_source_cases(input_path)
        prepared_cases, failures, duplicate_input_count = prepare_cases(
            raw_cases=raw_cases,
            start_index=args.start_index,
            limit=args.limit,
            allowed_target_groups=allowed_target_groups,
            allowed_scam_types=allowed_scam_types,
        )
    except Exception as exc:
        print(f"[import] 初始化失败: {exc}", file=sys.stderr)
        return 1

    remote_existing_keys: Set[Tuple[str, str, str]] = set()
    runtime_token = ""
    if args.skip_existing:
        try:
            runtime_token = resolve_token(api_base, args)
            remote_existing_keys = fetch_existing_remote_keys(api_base, runtime_token, args.timeout)
        except Exception as exc:
            print(f"[import] 拉取已有案件失败: {exc}", file=sys.stderr)
            return 1

    local_existing_skip = 0
    if remote_existing_keys:
        filtered_cases: List[Dict[str, Any]] = []
        for item in prepared_cases:
            payload = item["payload"]
            if build_remote_key(payload) in remote_existing_keys:
                local_existing_skip += 1
                continue
            filtered_cases.append(item)
        prepared_cases = filtered_cases

    print(
        "[import] 源记录: {}，预校验失败: {}，输入去重跳过: {}，远端已存在跳过: {}，待导入: {}".format(
            len(raw_cases),
            len(failures),
            duplicate_input_count,
            local_existing_skip,
            len(prepared_cases),
        )
    )

    if args.dry_run:
        if failure_output_path and failures:
            write_failures(failure_output_path, failures)
            print(f"[import] dry-run 失败记录已导出: {failure_output_path}")
        print("[import] dry-run 完成，未实际调用 API")
        return 0 if not failures else 1

    try:
        token = runtime_token or resolve_token(api_base, args)
    except Exception as exc:
        print(f"[import] 参数错误: {exc}", file=sys.stderr)
        return 1

    success_count = 0
    runtime_skip_existing = 0
    for index, item in enumerate(prepared_cases, start=1):
        payload = item["payload"]
        source_index = item["source_index"]
        remote_key = build_remote_key(payload)

        if remote_key in remote_existing_keys:
            runtime_skip_existing += 1
            continue

        try:
            response = request_json(
                f"{api_base}/scam/case-library/cases",
                method="POST",
                token=token,
                timeout=args.timeout,
                payload=payload,
            )
            case_info = response.get("case") if isinstance(response, dict) else {}
            created_case_id = ""
            if isinstance(case_info, dict):
                created_case_id = normalize_text(case_info.get("case_id"))
            success_count += 1
            remote_existing_keys.add(remote_key)
            print(
                "[import] success {}/{} source_index={} case_id={} title={}".format(
                    index,
                    len(prepared_cases),
                    source_index,
                    created_case_id or "-",
                    payload["title"],
                )
            )
            if args.sleep_ms > 0:
                time.sleep(args.sleep_ms / 1000.0)
        except Exception as exc:
            failures.append(
                {
                    "source_index": source_index,
                    "title": payload.get("title", ""),
                    "reason": str(exc),
                    "normalized_payload": payload,
                }
            )
            print(
                f"[import] failed {index}/{len(prepared_cases)} source_index={source_index} "
                f"title={payload['title']} error={exc}",
                file=sys.stderr,
            )
            if args.fail_fast:
                break

    if failure_output_path and failures:
        write_failures(failure_output_path, failures)
        print(f"[import] 失败记录已导出: {failure_output_path}")

    print(
        "[import] 完成: success={}，runtime_skip_existing={}，failed={}".format(
            success_count,
            runtime_skip_existing,
            len(failures),
        )
    )
    return 0 if not failures else 1


if __name__ == "__main__":
    sys.exit(main())