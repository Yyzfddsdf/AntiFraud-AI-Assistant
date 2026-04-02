import zipfile
import re
import os

doc_path = 'd:/Work/AntiFraud-AI-Assistant/competition_docs/专利/文档/反诈卫士（Sensible AI）_程序设计说明书_V1.0.docx'
out_path = 'd:/Work/AntiFraud-AI-Assistant/tmp/doc_debug.txt'

if os.path.exists(doc_path):
    with zipfile.ZipFile(doc_path, 'r') as zf:
        doc_xml = zf.read('word/document.xml').decode('utf-8')
        with open(out_path, 'w', encoding='utf-8') as f:
            f.write(doc_xml)
    print(f"Successfully wrote {out_path}")
else:
    print(f"File not found: {doc_path}")
