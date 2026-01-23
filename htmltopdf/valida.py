import json
import fitz  # PyMuPDF

def draw_anchors_on_pdf(pdf_path="contract.pdf", anchors_path="anchors.json", output_path="out_marked.pdf"):
    """
    在 PDF 文件上根据 anchors.json 中的坐标绘制矩形框
    
    参数:
        pdf_path: 输入的 PDF 文件路径
        anchors_path: anchors.json 文件路径
        output_path: 输出的标注后的 PDF 文件路径
    """
    # 读取 anchors.json
    with open(anchors_path, 'r', encoding='utf-8') as f:
        anchors = json.load(f)
    
    # 打开 PDF 文件
    doc = fitz.open(pdf_path)
    
    # 像素到点的转换比例
    # 假设浏览器渲染时使用的是 96 DPI
    # PDF 使用 72 DPI
    # 转换比例 = 72 / 96 = 0.75
    px_to_pt = 0.75
    
    # 为每个锚点绘制矩形框
    for anchor in anchors:
        page_num = anchor['page'] - 1  # PDF 页码从 0 开始
        
        if page_num >= len(doc):
            print(f"警告: 页码 {anchor['page']} 超出 PDF 总页数")
            continue
        
        page = doc[page_num]
        
        # 转换坐标（从 px 到 pt）
        x = anchor['x'] * px_to_pt
        y = anchor['y'] * px_to_pt
        width = anchor['width'] * px_to_pt
        height = anchor['height'] * px_to_pt
        
        # 创建矩形
        rect = fitz.Rect(x, y, x + width, y + height)
        
        # 绘制矩形框
        # 使用红色边框，2pt 宽度，无填充
        page.draw_rect(rect, color=(1, 0, 0), width=2)
        
        # 在矩形上方添加标签
        text_rect = fitz.Rect(x, y - 15, x + 200, y)
        page.insert_textbox(
            text_rect,
            anchor['name'],
            fontsize=10,
            color=(1, 0, 0),
            align=0
        )
    
    # 保存标注后的 PDF
    doc.save(output_path)
    doc.close()
    
    print(f"✅ 标注完成！输出文件: {output_path}")
    print(f"📊 共标注了 {len(anchors)} 个锚点")


if __name__ == "__main__":
    draw_anchors_on_pdf()
