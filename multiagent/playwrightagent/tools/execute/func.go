package execute

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // 导入支持 JPEG 格式的解码器
	_ "image/png"  // 导入支持 PNG 格式的解码器
	"log"

	"github.com/canfire/godemo/multiagent/playwrightagent/global"
	"github.com/canfire/godemo/multiagent/util/code"
	"github.com/canfire/godemo/multiagent/util/model"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

const (
	recognizePrompt = `用户需求：{{ query }}

请在图片页面上找出符合用户需求的操作区域，并给出操作区域的坐标，坐标要准确的覆盖用户可能操作的区域。

比如获取搜索结果就不要框选到搜索框的位置，但是要框选到所有搜索结果内容

当前图片的宽度为{{ width }}，高度为{{ height }}

输出格式,点(x1,y1)为区域左上角点，点(x2,y2)为区域右下角点
{
"x1":xxx,
"y1":xxx,
"x2":xxxx,
"y2":xxxx
}

只需要给出对应坐标的对应json结构即可，不要返回其他内容`

	operationPrompt = `## 下列是一组数据，内容为用户当前网页上的可操作元素
用户输入字段解释：
id字段为元素的id
tag字段为元素的类型例如 a，button，input等
text字段为元素的文本内容
Placeholder字段为元素的提示文字内容

## 请根据用户指令以及当其页面页面元素，生成对应的操作指令
- 用户指令：{{query}}
- 页面元素：{{elInfos}}
返回指令的格式为
[
    {
        "element_id": 1,
        "operation": "", // Click 点击 Fill 输入
        "content":"" // 如果为输入框则为输入框的值 
    }
     {
        "element_id": 20,
        "operation": "", // Click 点击 Fill 输入
        "content":"" // 如果为输入框则为输入框的值 
    }
]
`
)

// 获取当前页面要操作的区域
// 获取对应区域元素
// 获取执行指令
// 执行

// recognize 识别区域
func recognize(ctx context.Context, imgbyte []byte, query string) (*Region, error) {
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: model.NewVLChatModel(ctx),
	})
	if err != nil {
		log.Fatalf("getMultimodalAgent NewChatModelAgent err: %v", err)
		return nil, err
	}
	// 解析图片数据
	img, _, err := image.DecodeConfig(bytes.NewReader(imgbyte)) // 或者使用 image.Decode 如果你需要完整的图片对象
	if err != nil {
		log.Fatalf("recognize ===== %v", err)
	}

	temp := &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			{
				Type: schema.ChatMessagePartTypeText,
				Text: recognizePrompt,
			},
			{
				Type: schema.ChatMessagePartTypeImageURL,
				Image: &schema.MessageInputImage{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: of(base64.StdEncoding.EncodeToString(imgbyte)),
						MIMEType:   "image/png",
						Extra: map[string]any{
							"vl_high_resolution_images": true,
						},
					},
					Detail: schema.ImageURLDetailHigh,
				},
			},
		},
	}
	values := map[string]interface{}{
		"query":  query,
		"width":  img.Width,
		"height": img.Height,
	}
	msg, err := temp.Format(ctx, values, schema.Jinja2)
	if err != nil {
		log.Fatalf("Recognize Format error: %v", err)
		return nil, err
	}
	res, err := agent.Generate(ctx, msg)
	if err != nil {
		log.Fatalf("Recognize Generate error: %v", err)
		return nil, err
	}
	var region Region
	err = json.Unmarshal([]byte(code.ExtractCode("json", res.Content)), &region)
	if err != nil {
		log.Fatalf("Recognize Unmarshal error: %v", err)
		return nil, err
	}
	// 换算坐标
	region.X1 = (region.X1 / 1000) * float64(img.Width)
	region.Y1 = (region.Y1 / 1000) * float64(img.Height)
	region.X2 = (region.X2 / 1000) * float64(img.Width)
	region.Y2 = (region.Y2 / 1000) * float64(img.Height)
	fmt.Printf("\n=======================\n%+v\n", region)
	return &region, nil
}

// 获取指令 元素
func ExtractPageElements(ctx context.Context, query string) ([]ElementInfo, error) {
	page, err := global.GetPage(ctx)
	if err != nil {
		log.Fatalf("failed to get page: %v", err)
		return nil, err
	}
	imgbyte, err := page.Screenshot()
	if err != nil {
		log.Fatalf("failed to get screenshot: %v", err)
		return nil, err
	}
	region, err := recognize(ctx, imgbyte, query)
	if err != nil {
		log.Fatalf("failed to recognize region: %v", err)
		return nil, err
	}
	// 执行JavaScript提取父容器及其子元素
	result, err := page.Evaluate(`(region) => {
	const elements = [];

	const left   = Math.min(region.x1, region.x2);
	const right  = Math.max(region.x1, region.x2);
	const top    = Math.min(region.y1, region.y2);
	const bottom = Math.max(region.y1, region.y2);

	const interactiveSelectors = [
		'button', 'a', 'input', 'textarea', 'select',
		'[onclick]', '[role="button"]', '[role="link"]'
	];

	const allElements = document.querySelectorAll(interactiveSelectors.join(','));

	function intersects(rect) {
		return !(
			rect.right < left ||
			rect.left > right ||
			rect.bottom < top ||
			rect.top > bottom
		);
	}

	// 找到区域内的所有元素
	const elementsInRegion = [];
	allElements.forEach(el => {
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);

		if (
			rect.width > 0 &&
			rect.height > 0 &&
			style.visibility !== 'hidden' &&
			style.display !== 'none' &&
			intersects(rect)
		) {
			elementsInRegion.push(el);
		}
	});

	if (elementsInRegion.length === 0) {
		return elements;
	}

	// 找到最近的共同父容器
	function findCommonAncestor(elements) {
		if (elements.length === 0) return null;
		if (elements.length === 1) return elements[0].parentElement;

		// 获取第一个元素的所有祖先
		const ancestors = [];
		let current = elements[0];
		while (current) {
			ancestors.push(current);
			current = current.parentElement;
		}

		// 从最近的祖先开始，找到包含所有元素的第一个祖先
		for (let ancestor of ancestors) {
			const containsAll = elements.every(el => ancestor.contains(el));
			if (containsAll && ancestor !== elements[0]) {
				return ancestor;
			}
		}

		return document.body;
	}

	const commonParent = findCommonAncestor(elementsInRegion);
	
	// 识别区域内元素的结构特征
	function getElementPattern(el) {
		return {
			tag: el.tagName.toLowerCase(),
			classes: Array.from(el.classList).sort().join(' '),
			role: el.getAttribute('role') || '',
			// 获取在父元素中的相对位置
			parentTag: el.parentElement?.tagName.toLowerCase() || ''
		};
	}

	// 分析区域内元素的共同模式
	const patterns = elementsInRegion.map(el => {
		let current = el;
		while (current && current !== commonParent) {
			current = current.parentElement;
		}
		// 找到在commonParent下的直接或间接子元素结构
		current = el;
		let depth = 0;
		while (current.parentElement && current.parentElement !== commonParent && depth < 5) {
			current = current.parentElement;
			depth++;
		}
		return current;
	});

	// 去重，找到所有同级的容器
	const containerSet = new Set(patterns);
	const containers = Array.from(containerSet);

	// 如果找到了重复的容器结构，获取父容器下所有相似的子容器
	let targetElements = [];
	if (containers.length > 0 && commonParent) {
		const firstContainer = containers[0];
		const containerTag = firstContainer.tagName.toLowerCase();
		const containerClasses = Array.from(firstContainer.classList);
		
		// 在commonParent中查找所有相似的容器
		const siblings = Array.from(commonParent.children);
		const similarContainers = siblings.filter(sibling => {
			if (sibling.tagName.toLowerCase() !== containerTag) return false;
			
			// 检查是否有相同的class
			const siblingClasses = Array.from(sibling.classList);
			const hasCommonClass = containerClasses.some(cls => 
				siblingClasses.includes(cls) && cls.length > 0
			);
			
			return hasCommonClass || containerClasses.length === 0;
		});

		// 如果找到多个相似容器，说明这是一个列表结构
		if (similarContainers.length > 1) {
			// 从所有相似容器中提取交互元素
			similarContainers.forEach(container => {
				const interactiveInContainer = container.querySelectorAll(interactiveSelectors.join(','));
				targetElements.push(...Array.from(interactiveInContainer));
			});
		} else {
			// 否则返回commonParent下的所有交互元素
			targetElements = Array.from(commonParent.querySelectorAll(interactiveSelectors.join(',')));
		}
	} else {
		targetElements = elementsInRegion;
	}

	// 生成最终结果
	function getXPath(element) {
		if (element.id) return '//*[@id="' + element.id + '"]';
		if (element === document.body) return '/html/body';

		let ix = 0;
		const siblings = element.parentNode?.childNodes || [];
		for (let i = 0; i < siblings.length; i++) {
			const sibling = siblings[i];
			if (sibling === element) {
				const parent = element.parentNode;
				return getXPath(parent) + '/' +
					element.tagName.toLowerCase() + '[' + (ix + 1) + ']';
			}
			if (sibling.nodeType === 1 && sibling.tagName === element.tagName) {
				ix++;
			}
		}
	}

	targetElements.forEach((el, idx) => {
		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);

		if (
			rect.width > 0 &&
			rect.height > 0 &&
			style.visibility !== 'hidden' &&
			style.display !== 'none'
		) {
			el.setAttribute('data-ai-index', idx);

			elements.push({
				tag: el.tagName.toLowerCase(),
				text: el.innerText?.trim().substring(0, 100) || el.value || '',
				selector: '[data-ai-index="' + idx + '"]',
				xpath: getXPath(el),
				css_id: el.id || '',
				class: el.className || '',
				name: el.name || '',
				type: el.type || '',
				href: el.href || '',
				placeholder: el.placeholder || '',
				rect: {
					x: rect.x,
					y: rect.y,
					width: rect.width,
					height: rect.height
				}
			});
		}
	});

	return elements;
}`, map[string]interface{}{
		"x1": region.X1,
		"y1": region.Y1,
		"x2": region.X2,
		"y2": region.Y2,
	})

	if err != nil {
		log.Fatalf("ExtractPageElements Evaluate err: %v", err)
		return nil, fmt.Errorf("failed to extract elements: %w", err)
	}

	// 解析结果
	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("ExtractPageElements json.Marshal err: %v", err)
		return nil, err
	}

	var elements []ElementInfo
	if err := json.Unmarshal(jsonData, &elements); err != nil {
		log.Fatalf("ExtractPageElements json.Unmarshal err: %v", err)
		return nil, err
	}

	// 设置ID
	for i := range elements {
		elements[i].ID = i + 1
	}

	return elements, nil
}

func GetOperation(ctx context.Context, elInfos []ElementInfo, query string) ([]*OperationReq, error) {
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: model.NewChatModel(ctx),
	})
	if err != nil {
		log.Fatalf("GetOperation getAgent err: %v", err)
		return nil, err
	}
	temp := &schema.Message{
		Role:    schema.User,
		Content: operationPrompt,
	}
	elInfoList := elInfoList{}
	for _, v := range elInfos {
		elInfoList = append(elInfoList, elInfo{
			ID:          v.ID,
			Tag:         v.Tag,
			Text:        v.Text,
			Placeholder: v.Placeholder,
		})
	}
	values := map[string]interface{}{
		"query":   query,
		"elInfos": elInfoList.string(),
	}
	msg, err := temp.Format(ctx, values, schema.Jinja2)
	if err != nil {
		log.Fatalf("Recognize Format error: %v", err)
		return nil, err
	}
	// fmt.printf("=======================", msg)
	fmt.Printf("messages:%+v", msg)
	res, err := agent.Generate(ctx, msg)
	if err != nil {
		log.Fatalf("Recognize Generate error: %v", err)
		return nil, err
	}
	println("=======================", res.Content)

	var operationReqs []*OperationReq
	err = json.Unmarshal([]byte(code.ExtractCode("json", res.Content)), &operationReqs)
	if err != nil {
		log.Fatalf("Recognize Unmarshal error: %v", err)
		return nil, err
	}

	return operationReqs, nil
}

// Execute 执行操作
func Execute(ctx context.Context, step []*OperationReq, elInfos []ElementInfo) error {
	page, err := global.GetPage(ctx)
	if err != nil {
		log.Fatalf("failed to get page: %v", err)
		return err
	}
	elMap := map[int]ElementInfo{}
	for _, elInfo := range elInfos {
		elMap[elInfo.ID] = elInfo
	}
	for _, v := range step {
		elInfo, ok := elMap[v.ElementID]
		if !ok {
			log.Fatalf("Execute elInfo not found: %d", v.ElementID)
			return fmt.Errorf("Execute elInfo not found: %d", v.ElementID)
		}
		switch v.Operation {
		case OperationClick:
			err := page.Locator(elInfo.XPath).Click()
			if err != nil {
				log.Fatalf("%s", err.Error())
				return err
			}
		case OperationFill:
			err := page.Locator(elInfo.XPath).Fill(v.Content)
			if err != nil {
				log.Fatalf("%s", err.Error())
				return err
			}
		}
	}
	return nil
}

// Region 操作区域
type Region struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

func of[T any](t T) *T {
	return &t
}

type ElementInfo struct {
	ID          int    `json:"id"`
	Tag         string `json:"tag"`
	Text        string `json:"text"`
	Selector    string `json:"selector"`
	XPath       string `json:"xpath"`
	CssID       string `json:"css_id,omitempty"`
	Class       string `json:"class,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Href        string `json:"href,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Rect        struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"rect"`
}

type elInfo struct {
	ID          int    `json:"id"`
	Tag         string `json:"tag"`
	Text        string `json:"text"`
	Placeholder string `json:"placeholder"`
}

type elInfoList []elInfo

func (e elInfoList) string() string {
	elbyte, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(elbyte)
}

type Operation string

const (
	OperationClick = "Click" // 点击
	OperationFill  = "Fill"  // 输入
)

type OperationReq struct {
	ElementID int       `json:"element_id"`
	Operation Operation `json:"operation"`
	Content   string    `json:"content"`
}
