import argparse
import datetime as dt
import os
import zipfile
from xml.sax.saxutils import escape


W_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
R_NS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
CP_NS = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
DC_NS = "http://purl.org/dc/elements/1.1/"
DCTERMS_NS = "http://purl.org/dc/terms/"
XSI_NS = "http://www.w3.org/2001/XMLSchema-instance"
VT_NS = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"


def parse_args():
    parser = argparse.ArgumentParser(description="Convert markdown-like text to DOCX.")
    parser.add_argument("--input", required=True, help="Source markdown path.")
    parser.add_argument("--output", required=True, help="Target docx path.")
    parser.add_argument("--title", required=True, help="Document title.")
    parser.add_argument("--header", required=True, help="Header text.")
    return parser.parse_args()


def qn(tag):
    return f"w:{tag}"


def rqn(tag):
    return f"r:{tag}"


def make_run(text, *, bold=False, size=24, font="宋体", center=False):
    if text is None:
        text = ""
    text = escape(text)
    bold_xml = "<w:b/><w:bCs/>" if bold else ""
    jc_xml = "<w:jc w:val=\"center\"/>" if center else ""
    preserve = " xml:space=\"preserve\"" if text.startswith(" ") or text.endswith(" ") else ""
    return (
        f"<w:r>"
        f"<w:rPr>"
        f"{bold_xml}"
        f"<w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"{font}\" w:cs=\"Times New Roman\"/>"
        f"<w:sz w:val=\"{size}\"/><w:szCs w:val=\"{size}\"/>"
        f"</w:rPr>"
        f"<w:t{preserve}>{text}</w:t>"
        f"</w:r>"
    ), jc_xml


def make_paragraph(text="", *, style=None, bold=False, size=24, spacing_before=0, spacing_after=0,
                   line=360, left=0, first_line=0, center=False, page_break_before=False):
    run_xml, center_xml = make_run(text, bold=bold, size=size, center=center)
    style_xml = f"<w:pStyle w:val=\"{style}\"/>" if style else ""
    indent_xml = ""
    if left or first_line:
        attrs = []
        if left:
            attrs.append(f"w:left=\"{left}\"")
        if first_line:
            attrs.append(f"w:firstLine=\"{first_line}\"")
        indent_xml = f"<w:ind {' '.join(attrs)}/>"
    page_break_xml = "<w:pageBreakBefore/>" if page_break_before else ""
    ppr = (
        f"<w:pPr>"
        f"{style_xml}"
        f"{center_xml}"
        f"<w:spacing w:before=\"{spacing_before}\" w:after=\"{spacing_after}\" w:line=\"{line}\" w:lineRule=\"exact\"/>"
        f"{indent_xml}"
        f"{page_break_xml}"
        f"</w:pPr>"
    )
    return f"<w:p>{ppr}{run_xml}</w:p>"


def make_blank_paragraph():
    return (
        "<w:p>"
        "<w:pPr><w:spacing w:before=\"0\" w:after=\"0\" w:line=\"360\" w:lineRule=\"exact\"/></w:pPr>"
        "<w:r><w:t></w:t></w:r>"
        "</w:p>"
    )


def iter_paragraphs(lines):
    paragraphs = []
    for idx, raw in enumerate(lines):
        line = raw.rstrip("\n").rstrip("\r")
        stripped = line.strip()

        if not stripped:
            paragraphs.append(make_blank_paragraph())
            continue

        if stripped == "---":
            paragraphs.append(make_blank_paragraph())
            continue

        if idx == 0 and stripped.startswith("# "):
            paragraphs.append(make_paragraph(
                stripped[2:].strip(),
                style="Title",
                bold=True,
                size=36,
                spacing_before=0,
                spacing_after=120,
                line=420,
                center=True,
            ))
            continue

        if stripped.startswith("# "):
            paragraphs.append(make_paragraph(
                stripped[2:].strip(),
                style="Heading1",
                bold=True,
                size=32,
                spacing_before=120,
                spacing_after=60,
                line=420,
            ))
            continue

        if stripped.startswith("## "):
            paragraphs.append(make_paragraph(
                stripped[3:].strip(),
                style="Heading1",
                bold=True,
                size=30,
                spacing_before=120,
                spacing_after=40,
                line=400,
            ))
            continue

        if stripped.startswith("### "):
            paragraphs.append(make_paragraph(
                stripped[4:].strip(),
                style="Heading2",
                bold=True,
                size=28,
                spacing_before=80,
                spacing_after=20,
                line=380,
            ))
            continue

        if stripped.startswith("#### "):
            paragraphs.append(make_paragraph(
                stripped[5:].strip(),
                style="Heading3",
                bold=True,
                size=26,
                spacing_before=40,
                spacing_after=20,
                line=360,
            ))
            continue

        if stripped.startswith("【") and stripped.endswith("】"):
            paragraphs.append(make_paragraph(
                stripped,
                bold=False,
                size=22,
                spacing_before=20,
                spacing_after=20,
                line=360,
            ))
            continue

        # Metadata lines near the top are centered.
        if idx < 8 and "：" in stripped and not stripped[:2].isdigit():
            paragraphs.append(make_paragraph(
                stripped,
                bold=False,
                size=24,
                spacing_before=0,
                spacing_after=0,
                line=360,
                center=True,
            ))
            continue

        first_line = 0
        left = 0
        if stripped.startswith(("- ", "* ")):
            stripped = "• " + stripped[2:].strip()
            left = 0
        elif stripped[:2].isdigit() and stripped[1:2] == ".":
            left = 0
        else:
            first_line = 0

        paragraphs.append(make_paragraph(
            stripped,
            bold=False,
            size=24,
            spacing_before=0,
            spacing_after=0,
            line=360,
            left=left,
            first_line=first_line,
        ))
    return paragraphs


def build_document_xml(lines):
    body = "".join(iter_paragraphs(lines))
    sect_pr = (
        "<w:sectPr>"
        "<w:headerReference w:type=\"default\" r:id=\"rIdHeader1\"/>"
        "<w:footerReference w:type=\"default\" r:id=\"rIdFooter1\"/>"
        "<w:pgSz w:w=\"11906\" w:h=\"16838\"/>"
        "<w:pgMar w:top=\"1440\" w:right=\"1440\" w:bottom=\"1440\" w:left=\"1440\" w:header=\"720\" w:footer=\"720\" w:gutter=\"0\"/>"
        "<w:cols w:space=\"425\"/>"
        "<w:docGrid w:linePitch=\"360\"/>"
        "</w:sectPr>"
    )
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:document xmlns:w=\"{W_NS}\" xmlns:r=\"{R_NS}\">"
        f"<w:body>{body}{sect_pr}</w:body>"
        "</w:document>"
    )


def build_styles_xml():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:styles xmlns:w=\"{W_NS}\">"
        "<w:docDefaults>"
        "<w:rPrDefault><w:rPr>"
        "<w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\" w:cs=\"Times New Roman\"/>"
        "<w:sz w:val=\"24\"/><w:szCs w:val=\"24\"/>"
        "<w:lang w:val=\"zh-CN\" w:eastAsia=\"zh-CN\" w:bidi=\"en-US\"/>"
        "</w:rPr></w:rPrDefault>"
        "<w:pPrDefault><w:pPr><w:spacing w:before=\"0\" w:after=\"0\" w:line=\"360\" w:lineRule=\"exact\"/></w:pPr></w:pPrDefault>"
        "</w:docDefaults>"
        "<w:style w:type=\"paragraph\" w:default=\"1\" w:styleId=\"Normal\">"
        "<w:name w:val=\"Normal\"/>"
        "<w:qFormat/>"
        "<w:rPr><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\" w:cs=\"Times New Roman\"/><w:sz w:val=\"24\"/><w:szCs w:val=\"24\"/></w:rPr>"
        "</w:style>"
        "<w:style w:type=\"paragraph\" w:styleId=\"Title\">"
        "<w:name w:val=\"Title\"/>"
        "<w:basedOn w:val=\"Normal\"/>"
        "<w:qFormat/>"
        "<w:pPr><w:jc w:val=\"center\"/><w:spacing w:before=\"0\" w:after=\"120\" w:line=\"420\" w:lineRule=\"exact\"/></w:pPr>"
        "<w:rPr><w:b/><w:bCs/><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"黑体\" w:cs=\"Times New Roman\"/><w:sz w:val=\"36\"/><w:szCs w:val=\"36\"/></w:rPr>"
        "</w:style>"
        "<w:style w:type=\"paragraph\" w:styleId=\"Heading1\">"
        "<w:name w:val=\"Heading 1\"/>"
        "<w:basedOn w:val=\"Normal\"/>"
        "<w:qFormat/>"
        "<w:pPr><w:spacing w:before=\"120\" w:after=\"40\" w:line=\"400\" w:lineRule=\"exact\"/></w:pPr>"
        "<w:rPr><w:b/><w:bCs/><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"黑体\" w:cs=\"Times New Roman\"/><w:sz w:val=\"30\"/><w:szCs w:val=\"30\"/></w:rPr>"
        "</w:style>"
        "<w:style w:type=\"paragraph\" w:styleId=\"Heading2\">"
        "<w:name w:val=\"Heading 2\"/>"
        "<w:basedOn w:val=\"Normal\"/>"
        "<w:qFormat/>"
        "<w:pPr><w:spacing w:before=\"80\" w:after=\"20\" w:line=\"380\" w:lineRule=\"exact\"/></w:pPr>"
        "<w:rPr><w:b/><w:bCs/><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"黑体\" w:cs=\"Times New Roman\"/><w:sz w:val=\"28\"/><w:szCs w:val=\"28\"/></w:rPr>"
        "</w:style>"
        "<w:style w:type=\"paragraph\" w:styleId=\"Heading3\">"
        "<w:name w:val=\"Heading 3\"/>"
        "<w:basedOn w:val=\"Normal\"/>"
        "<w:qFormat/>"
        "<w:pPr><w:spacing w:before=\"40\" w:after=\"20\" w:line=\"360\" w:lineRule=\"exact\"/></w:pPr>"
        "<w:rPr><w:b/><w:bCs/><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"黑体\" w:cs=\"Times New Roman\"/><w:sz w:val=\"26\"/><w:szCs w:val=\"26\"/></w:rPr>"
        "</w:style>"
        "</w:styles>"
    )


def build_header_xml(header_text):
    text = escape(header_text)
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:hdr xmlns:w=\"{W_NS}\" xmlns:r=\"{R_NS}\">"
        "<w:p>"
        "<w:pPr><w:jc w:val=\"center\"/></w:pPr>"
        "<w:r><w:rPr><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\"/><w:sz w:val=\"21\"/><w:szCs w:val=\"21\"/></w:rPr>"
        f"<w:t>{text}</w:t></w:r>"
        "</w:p>"
        "</w:hdr>"
    )


def build_footer_xml():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:ftr xmlns:w=\"{W_NS}\" xmlns:r=\"{R_NS}\">"
        "<w:p>"
        "<w:pPr><w:jc w:val=\"center\"/></w:pPr>"
        "<w:r><w:rPr><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\"/><w:sz w:val=\"21\"/><w:szCs w:val=\"21\"/></w:rPr><w:t>第 </w:t></w:r>"
        "<w:r><w:fldChar w:fldCharType=\"begin\"/></w:r>"
        "<w:r><w:instrText xml:space=\"preserve\"> PAGE </w:instrText></w:r>"
        "<w:r><w:fldChar w:fldCharType=\"separate\"/></w:r>"
        "<w:r><w:t>1</w:t></w:r>"
        "<w:r><w:fldChar w:fldCharType=\"end\"/></w:r>"
        "<w:r><w:rPr><w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\"/><w:sz w:val=\"21\"/><w:szCs w:val=\"21\"/></w:rPr><w:t> 页</w:t></w:r>"
        "</w:p>"
        "</w:ftr>"
    )


def build_document_rels():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\">"
        "<Relationship Id=\"rIdStyles\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles\" Target=\"styles.xml\"/>"
        "<Relationship Id=\"rIdSettings\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings\" Target=\"settings.xml\"/>"
        "<Relationship Id=\"rIdHeader1\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/header\" Target=\"header1.xml\"/>"
        "<Relationship Id=\"rIdFooter1\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer\" Target=\"footer1.xml\"/>"
        "</Relationships>"
    )


def build_root_rels():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\">"
        "<Relationship Id=\"rId1\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument\" Target=\"word/document.xml\"/>"
        "<Relationship Id=\"rId2\" Type=\"http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties\" Target=\"docProps/core.xml\"/>"
        "<Relationship Id=\"rId3\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties\" Target=\"docProps/app.xml\"/>"
        "</Relationships>"
    )


def build_content_types():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Types xmlns=\"http://schemas.openxmlformats.org/package/2006/content-types\">"
        "<Default Extension=\"rels\" ContentType=\"application/vnd.openxmlformats-package.relationships+xml\"/>"
        "<Default Extension=\"xml\" ContentType=\"application/xml\"/>"
        "<Override PartName=\"/word/document.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml\"/>"
        "<Override PartName=\"/word/styles.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml\"/>"
        "<Override PartName=\"/word/settings.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml\"/>"
        "<Override PartName=\"/word/header1.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml\"/>"
        "<Override PartName=\"/word/footer1.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml\"/>"
        "<Override PartName=\"/docProps/core.xml\" ContentType=\"application/vnd.openxmlformats-package.core-properties+xml\"/>"
        "<Override PartName=\"/docProps/app.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.extended-properties+xml\"/>"
        "</Types>"
    )


def build_settings():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:settings xmlns:w=\"{W_NS}\">"
        "<w:zoom w:percent=\"100\"/>"
        "<w:characterSpacingControl w:val=\"doNotCompress\"/>"
        "<w:compat/>"
        "</w:settings>"
    )


def build_core_xml(title):
    now = dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    title = escape(title)
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<cp:coreProperties xmlns:cp=\"{CP_NS}\" xmlns:dc=\"{DC_NS}\" xmlns:dcterms=\"{DCTERMS_NS}\" xmlns:dcmitype=\"http://purl.org/dc/dcmitype/\" xmlns:xsi=\"{XSI_NS}\">"
        f"<dc:title>{title}</dc:title>"
        "<dc:creator>Codex</dc:creator>"
        "<cp:lastModifiedBy>Codex</cp:lastModifiedBy>"
        f"<dcterms:created xsi:type=\"dcterms:W3CDTF\">{now}</dcterms:created>"
        f"<dcterms:modified xsi:type=\"dcterms:W3CDTF\">{now}</dcterms:modified>"
        "</cp:coreProperties>"
    )


def build_app_xml():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Properties xmlns=\"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties\" "
        f"xmlns:vt=\"{VT_NS}\">"
        "<Application>Microsoft Office Word</Application>"
        "<DocSecurity>0</DocSecurity>"
        "<ScaleCrop>false</ScaleCrop>"
        "<Company></Company>"
        "<LinksUpToDate>false</LinksUpToDate>"
        "<SharedDoc>false</SharedDoc>"
        "<HyperlinksChanged>false</HyperlinksChanged>"
        "<AppVersion>16.0000</AppVersion>"
        "</Properties>"
    )


def ensure_parent(path):
    parent = os.path.dirname(path)
    if parent:
        os.makedirs(parent, exist_ok=True)


def main():
    args = parse_args()
    ensure_parent(args.output)

    with open(args.input, "r", encoding="utf-8") as f:
        lines = f.readlines()

    with zipfile.ZipFile(args.output, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        zf.writestr("[Content_Types].xml", build_content_types())
        zf.writestr("_rels/.rels", build_root_rels())
        zf.writestr("docProps/core.xml", build_core_xml(args.title))
        zf.writestr("docProps/app.xml", build_app_xml())
        zf.writestr("word/document.xml", build_document_xml(lines))
        zf.writestr("word/styles.xml", build_styles_xml())
        zf.writestr("word/settings.xml", build_settings())
        zf.writestr("word/header1.xml", build_header_xml(args.header))
        zf.writestr("word/footer1.xml", build_footer_xml())
        zf.writestr("word/_rels/document.xml.rels", build_document_rels())

    print(args.output)


if __name__ == "__main__":
    main()
