import argparse
import datetime as dt
import os
import zipfile
import re
from xml.sax.saxutils import escape


W_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
R_NS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
CP_NS = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
DC_NS = "http://purl.org/dc/elements/1.1/"
DCTERMS_NS = "http://purl.org/dc/terms/"
XSI_NS = "http://www.w3.org/2001/XMLSchema-instance"
VT_NS = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"
WP_NS = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
A_NS = "http://schemas.openxmlformats.org/drawingml/2006/main"
PIC_NS = "http://schemas.openxmlformats.org/drawingml/2006/picture"
ASVG_NS = "http://schemas.microsoft.com/office/drawing/2016/SVG/main"


def parse_args():
    parser = argparse.ArgumentParser(description="Convert markdown-like text to DOCX with SVG/PNG/Table support.")
    parser.add_argument("--input", required=True, help="Source markdown path.")
    parser.add_argument("--output", required=True, help="Target docx path.")
    parser.add_argument("--title", required=True, help="Document title.")
    parser.add_argument("--header", required=True, help="Header text.")
    return parser.parse_args()


def parse_inline_markdown(text):
    parts = []
    # Strip spaces
    text = text.strip()
    tokens = re.split(r"(\*\*.*?\*\*)", text)
    for token in tokens:
        if token.startswith("**") and token.endswith("**"):
            parts.append({"text": token[2:-2], "bold": True})
        else:
            parts.append({"text": token, "bold": False})
    return parts


def make_run_xml(text, *, bold=False, size=24, font="宋体"):
    if not text:
        return ""
    text_esc = escape(text)
    bold_xml = "<w:b/><w:bCs/>" if bold else ""
    preserve = " xml:space=\"preserve\"" if text.startswith(" ") or text.endswith(" ") else ""
    return (
        f"<w:r>"
        f"<w:rPr>"
        f"{bold_xml}"
        f"<w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"{font}\" w:cs=\"Times New Roman\"/>"
        f"<w:sz w:val=\"{size}\"/><w:szCs w:val=\"{size}\"/>"
        f"</w:rPr>"
        f"<w:t{preserve}>{text_esc}</w:t>"
        f"</w:r>"
    )


def make_p_content(text, size=24, font="宋体"):
    chunks = parse_inline_markdown(text)
    xml = ""
    for c in chunks:
        xml += make_run_xml(c["text"], bold=c["bold"], size=size, font=font)
    return xml


def make_paragraph(text="", *, bold=False, size=24, spacing_before=0, spacing_after=0,
                   line=360, left=0, first_line=0, center=False):
    jc_xml = f"<w:jc w:val=\"{'center' if center else 'left'}\"/>"
    indent_xml = ""
    if left or first_line:
        attrs = []
        if left: attrs.append(f"w:left=\"{left}\"")
        if first_line: attrs.append(f"w:firstLine=\"{first_line}\"")
        indent_xml = f"<w:ind {' '.join(attrs)}/>"
        
    content_xml = make_p_content(text, size=size)
    return (
        f"<w:p>"
        f"<w:pPr>"
        f"<w:spacing w:before=\"{spacing_before}\" w:after=\"{spacing_after}\" w:line=\"{line}\" w:lineRule=\"exact\"/>"
        f"{indent_xml}"
        f"{jc_xml}"
        f"</w:pPr>"
        f"{content_xml}"
        f"</w:p>"
    )


def make_heading(text, level, first_h1=False):
    if first_h1:
        return make_paragraph(text, bold=True, size=44, center=True, spacing_before=400, spacing_after=400)
    sizes = {1: 32, 2: 28, 3: 24, 4: 24}
    size = sizes.get(level, 24)
    # Headings: no first line indent, some spacing
    return make_paragraph(text, bold=True, size=size, spacing_before=240, spacing_after=120, line=400, first_line=0)


def make_table(rows):
    # rows is list of lists
    tbl_pr = (
        "<w:tblPr>"
        "<w:tblW w:w=\"0\" w:type=\"auto\"/>"
        "<w:jc w:val=\"center\"/>"
        "<w:tblBorders>"
        "<w:top w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "<w:left w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "<w:bottom w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "<w:right w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "<w:insideH w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "<w:insideV w:val=\"single\" w:sz=\"4\" w:space=\"0\" w:color=\"auto\"/>"
        "</w:tblBorders>"
        "</w:tblPr>"
    )
    
    tr_xml = ""
    for idx, row in enumerate(rows):
        tc_xml = ""
        for cell in row:
            # Use smaller font for tables
            content = make_p_content(cell.strip(), size=21)
            tc_xml += (
                f"<w:tc><w:tcPr><w:tcW w:w=\"0\" w:type=\"auto\"/></w:tcPr>"
                f"<w:p><w:pPr><w:spacing w:line=\"240\" w:lineRule=\"auto\"/><w:jc w:val=\"left\"/></w:pPr>{content}</w:p>"
                f"</w:tc>"
            )
        tr_xml += f"<w:tr>{tc_xml}</w:tr>"
        
    # Wrap table in spacing paragraphs
    before = "<w:p><w:pPr><w:spacing w:after=\"120\"/></w:pPr></w:p>"
    after = "<w:p><w:pPr><w:spacing w:before=\"120\"/></w:pPr></w:p>"
    return f"{before}<w:tbl>{tbl_pr}{tr_xml}</w:tbl>{after}"


def make_image_drawing(rId_png, rId_svg, alt_text, width_emu=5486400, height_emu=3429000):
    doc_id = "".join(filter(str.isdigit, rId_png)) or "1"
    alt_text_esc = escape(alt_text or "image")
    return (
        f"<w:p><w:pPr><w:spacing w:before=\"240\" w:after=\"240\"/><w:jc w:val=\"center\"/></w:pPr>"
        f"<w:r><w:rPr><w:noProof/></w:rPr>"
        f"<w:drawing><wp:inline distT=\"0\" distB=\"0\" distL=\"114300\" distR=\"114300\">"
        f"<wp:extent cx=\"{width_emu}\" cy=\"{height_emu}\"/>"
        f"<wp:effectExtent l=\"0\" t=\"0\" r=\"0\" b=\"0\"/>"
        f"<wp:docPr id=\"{doc_id}\" name=\"图片 {doc_id}\" descr=\"{alt_text_esc}\"/>"
        f"<wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a=\"{A_NS}\" noChangeAspect=\"1\"/></wp:cNvGraphicFramePr>"
        f"<a:graphic xmlns:a=\"{A_NS}\">"
        f"<a:graphicData uri=\"{PIC_NS}\">"
        f"<pic:pic xmlns:pic=\"{PIC_NS}\">"
        f"<pic:nvPicPr>"
        f"<pic:cNvPr id=\"{doc_id}\" name=\"图片 {doc_id}\" descr=\"{alt_text_esc}\"/>"
        f"<pic:cNvPicPr><a:picLocks noChangeAspect=\"1\"/></pic:cNvPicPr>"
        f"</pic:nvPicPr>"
        f"<pic:blipFill>"
        f"<a:blip r:embed=\"{rId_png}\">"
        f"<a:extLst>"
        f"<a:ext uri=\"{{96DAC541-7B7A-43D3-8B79-37D633B846F1}}\">"
        f"<asvg:svgBlip xmlns:asvg=\"{ASVG_NS}\" r:embed=\"{rId_svg}\"/>"
        f"</a:ext>"
        f"</a:extLst>"
        f"</a:blip>"
        f"<a:stretch><a:fillRect/></a:stretch>"
        f"</pic:blipFill>"
        f"<pic:spPr>"
        f"<a:xfrm><a:off x=\"0\" y=\"0\"/><a:ext cx=\"{width_emu}\" cy=\"{height_emu}\"/></a:xfrm>"
        f"<a:prstGeom prst=\"rect\"><a:avLst/></a:prstGeom>"
        f"</pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>"
    )


def iter_paragraphs(lines, image_map):
    paragraphs = []
    img_re = re.compile(r"!\[(.*?)\]\((.*?)\)")

    table_buffer = []
    first_h1 = True
    
    def flush_table():
        if table_buffer:
            rows = []
            for t_line in table_buffer:
                if t_line.strip() and not re.match(r"^\|?\s*:?-+:?\s*\|", t_line.strip()):
                    actual_cells = t_line.strip("|").split("|")
                    rows.append(actual_cells)
            if rows:
                paragraphs.append(make_table(rows))
            table_buffer.clear()

    for line in lines:
        stripped = line.strip()
        
        # Table detection
        if stripped.startswith("|") and "|" in stripped:
            table_buffer.append(stripped)
            continue
        else:
            flush_table()

        if not stripped:
            continue
            
        img_match = img_re.search(stripped)
        if img_match:
            alt_text = img_match.group(1)
            img_path = img_match.group(2)
            if img_path in image_map:
                rIds = image_map[img_path]
                paragraphs.append(make_image_drawing(rIds["png"], rIds["svg"], alt_text))
                continue

        header_match = re.match(r"^(#{1,6})\s+(.*)", stripped)
        if header_match:
            level = len(header_match.group(1))
            text = header_match.group(2)
            if level == 1 and first_h1:
                paragraphs.append(make_heading(text, level, first_h1=True))
                first_h1 = False
            else:
                paragraphs.append(make_heading(text, level))
            continue
            
        if stripped == "---": continue

        if "：" in stripped and len(stripped) < 100 and not stripped.startswith("**"):
            # Metadata: no indent
            paragraphs.append(make_paragraph(stripped, first_line=0))
            continue

        if re.match(r"^\d+\.\s+.*", stripped) or stripped.startswith("- "):
            # List items: indent left 480, no first line
            paragraphs.append(make_paragraph(stripped, left=480, first_line=0))
        else:
            # Normal paragraph: first line 480
            paragraphs.append(make_paragraph(stripped, first_line=480))
            
    flush_table()
    return paragraphs


def build_content_types(has_svg=False):
    svg_str = "<Default Extension=\"svg\" ContentType=\"image/svg+xml\"/>" if has_svg else ""
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Types xmlns=\"http://schemas.openxmlformats.org/package/2006/content-types\">"
        "<Default Extension=\"rels\" ContentType=\"application/vnd.openxmlformats-package.relationships+xml\"/>"
        "<Default Extension=\"xml\" ContentType=\"application/xml\"/>"
        "<Default Extension=\"png\" ContentType=\"image/png\"/>"
        f"{svg_str}"
        "<Override PartName=\"/word/document.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml\"/>"
        "<Override PartName=\"/word/styles.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml\"/>"
        "<Override PartName=\"/word/settings.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml\"/>"
        "<Override PartName=\"/word/header1.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml\"/>"
        "<Override PartName=\"/word/footer1.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml\"/>"
        "<Override PartName=\"/docProps/core.xml\" ContentType=\"application/vnd.openxmlformats-package.core-properties+xml\"/>"
        "<Override PartName=\"/docProps/app.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.extended-properties+xml\"/>"
        "</Types>"
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


def build_core_xml(title):
    now = dt.datetime.utcnow().isoformat() + "Z"
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<cp:coreProperties xmlns:cp=\"{CP_NS}\" xmlns:dc=\"{DC_NS}\" xmlns:dcterms=\"{DCTERMS_NS}\" xmlns:xsi=\"{XSI_NS}\">"
        f"<dc:title>{escape(title)}</dc:title>"
        "<dc:creator>System Agent</dc:creator>"
        f"<dcterms:created xsi:type=\"dcterms:W3CDTF\">{now}</dcterms:created>"
        "</cp:coreProperties>"
    )


def build_app_xml():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Properties xmlns=\"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties\">"
        "<Application>Markdown to DOCX Tool</Application>"
        "</Properties>"
    )


def build_styles_xml():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:styles xmlns:w=\"{W_NS}\">"
        "<w:docDefaults><w:rPrDefault><w:rPr>"
        "<w:rFonts w:ascii=\"Times New Roman\" w:hAnsi=\"Times New Roman\" w:eastAsia=\"宋体\" w:cs=\"Times New Roman\"/>"
        "<w:sz w:val=\"24\"/><w:szCs w:val=\"24\"/>"
        "</w:rPr></w:rPrDefault></w:docDefaults>"
        "</w:styles>"
    )


def build_settings():
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:settings xmlns:w=\"{W_NS}\">"
        "<w:displayBackgroundShape/>"
        "</w:settings>"
    )


def build_header_xml(text):
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:hdr xmlns:w=\"{W_NS}\">"
        "<w:p>"
        "<w:pPr><w:jc w:val=\"center\"/><w:pBdr><w:bottom w:val=\"single\" w:sz=\"4\" w:space=\"1\" w:color=\"auto\"/></w:pBdr></w:pPr>"
        f"{make_run_xml(text, size=18)}"
        "</w:p>"
        "</hdr>"
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


def build_document_xml(lines, image_map):
    paragraphs = iter_paragraphs(lines, image_map)
    body_content = "".join(paragraphs)
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        f"<w:document xmlns:w=\"{W_NS}\" xmlns:r=\"{R_NS}\" xmlns:wp=\"{WP_NS}\" xmlns:a=\"{A_NS}\" xmlns:pic=\"{PIC_NS}\">"
        f"<w:body>{body_content}"
        "<w:sectPr>"
        "<w:headerReference w:type=\"default\" r:id=\"rIdHeader\"/>"
        "<w:footerReference w:type=\"default\" r:id=\"rIdFooter\"/>"
        "<w:pgSz w:w=\"11906\" w:h=\"16838\"/>"
        "<w:pgMar w:top=\"1440\" w:right=\"1800\" w:bottom=\"1440\" w:left=\"1800\" w:header=\"851\" w:footer=\"992\" w:gutter=\"0\"/>"
        "</w:sectPr>"
        "</w:body></w:document>"
    )


def build_document_rels(image_map):
    rels = [
        f"<Relationship Id=\"rIdHeader\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/header\" Target=\"header1.xml\"/>",
        f"<Relationship Id=\"rIdFooter\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer\" Target=\"footer1.xml\"/>",
        f"<Relationship Id=\"rIdStyles\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles\" Target=\"styles.xml\"/>",
        f"<Relationship Id=\"rIdSettings\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings\" Target=\"settings.xml\"/>",
    ]
    for path, rIds in image_map.items():
        svg_filename = f"image_{rIds['svg'][5:]}.svg"
        png_filename = f"image_{rIds['png'][5:]}.png"
        rels.append(f"<Relationship Id=\"{rIds['png']}\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/image\" Target=\"media/{png_filename}\"/>")
        rels.append(f"<Relationship Id=\"{rIds['svg']}\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/image\" Target=\"media/{svg_filename}\"/>")
        
    return (
        "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>"
        "<Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\">"
        + "".join(rels) +
        "</Relationships>"
    )


def main():
    args = parse_args()
    if not os.path.exists(args.input):
        print(f"Error: Input file {args.input} not found.")
        return
    with open(args.input, "r", encoding="utf-8") as f:
        lines = f.readlines()
    image_map = {}
    img_re = re.compile(r"!\[(.*?)\]\((.*?)\)")
    img_count = 1
    input_dir = os.path.dirname(os.path.abspath(args.input))
    for line in lines:
        for match in img_re.finditer(line):
            path = match.group(2)
            if path not in image_map:
                image_map[path] = {
                    "svg": f"rIdSv{img_count:02d}",
                    "png": f"rIdPn{img_count:02d}"
                }
                img_count += 1
    with zipfile.ZipFile(args.output, "w") as zf:
        zf.writestr("[Content_Types].xml", build_content_types(has_svg=True))
        zf.writestr("_rels/.rels", build_root_rels())
        zf.writestr("docProps/core.xml", build_core_xml(args.title))
        zf.writestr("docProps/app.xml", build_app_xml())
        zf.writestr("word/document.xml", build_document_xml(lines, image_map))
        zf.writestr("word/styles.xml", build_styles_xml())
        zf.writestr("word/settings.xml", build_settings())
        zf.writestr("word/header1.xml", build_header_xml(args.header))
        zf.writestr("word/footer1.xml", build_footer_xml())
        zf.writestr("word/_rels/document.xml.rels", build_document_rels(image_map))
        for path, rIds in image_map.items():
            svg_abs_path = os.path.normpath(os.path.join(input_dir, path))
            png_rel_path = path.replace("/svg/", "/png/").replace(".svg", ".png")
            png_abs_path = os.path.normpath(os.path.join(input_dir, png_rel_path))
            if os.path.exists(svg_abs_path):
                svg_filename = f"image_{rIds['svg'][5:]}.svg"
                with open(svg_abs_path, "rb") as img_f:
                    zf.writestr(f"word/media/{svg_filename}", img_f.read())
            if os.path.exists(png_abs_path):
                png_filename = f"image_{rIds['png'][5:]}.png"
                with open(png_abs_path, "rb") as img_f:
                    zf.writestr(f"word/media/{png_filename}", img_f.read())
    print(f"Successfully generated {args.output}")


if __name__ == "__main__":
    main()
