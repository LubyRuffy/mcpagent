package config

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cloudwego/eino-ext/components/tool/sequentialthinking"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// 定义内置工具的常量
const (
	// InnerServerName 是内置工具服务器的名称
	InnerServerName = "inner"

	// SequentialThinkingToolName 是顺序思考工具的名称
	SequentialThinkingToolName = "sequentialthinking"
)

// Result defines a search query result type.
type Result struct {
	Title string
	Info  string
	Ref   string
}

const DuckDuckGoURL = "https://html.duckduckgo.com/html/"

type ddgWebSearchInput struct {
	Query   string
	Page    int
	Timeout time.Duration
}

func ddgWebSearch(ctx context.Context, input ddgWebSearchInput) (urls []Result, err error) {
	c := http.Client{
		Timeout: input.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	serializedPage := func() string {
		switch input.Page {
		case 0, 1:
			return ""
		case 2:
			return "29"
		default:
			return strconv.Itoa(input.Page*50 + 29)
		}
	}()

	req, err := http.NewRequest(http.MethodPost, DuckDuckGoURL,
		bytes.NewReader([]byte("b=&s="+serializedPage+"&q="+input.Query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://html.duckduckgo.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Origin", "https://html.duckduckgo.com")
	req.Header.Set("Cookie", "kl=wt-wt")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("new document error: %w", err)
	}

	results := []Result{}
	sel := doc.Find(".web-result")

	for i := range sel.Nodes {
		node := sel.Eq(i)
		titleNode := node.Find(".result__a")

		info := node.Find(".result__snippet").Text()
		title := titleNode.Text()
		ref := ""

		if len(titleNode.Nodes) > 0 && len(titleNode.Nodes[0].Attr) > 2 {
			ref, err = url.QueryUnescape(
				strings.TrimPrefix(
					titleNode.Nodes[0].Attr[2].Val,
					"/l/?kh=-1&uddg=",
				),
			)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, Result{title, info, ref})
	}

	return results, nil
}

func newWebSearchTool(ctx context.Context, proxy string) (tool.InvokableTool, error) {
	tool, err := utils.InferTool(
		"search",
		"search web for information by duckduckgo",
		ddgWebSearch,
	)
	if err != nil {
		return nil, err
	}

	return tool, nil
}

// GetInternalTools 返回所有内置工具列表
func GetInternalTools(ctx context.Context, proxy string) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	// 创建顺序思考工具并进行自定义
	seqThinking, err := sequentialthinking.NewTool()
	if err != nil {
		return nil, fmt.Errorf("创建顺序思考工具失败: %w", err)
	}

	// 创建DuckDuckGo工具
	// 使用eino-ext/components/tool/duckduckgo，目前官方的还有bug，先自己实现
	// ddgConfig := &ddgsearch.Config{}
	// if proxy != "" {
	// 	ddgConfig.Proxy = proxy
	// 	ddgConfig.Headers = map[string]string{
	// 		"Content-Type":    "application/x-www-form-urlencoded",
	// 		"Referer":         "https://html.duckduckgo.com/",
	// 		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	// 		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	// 		"Accept-Language": "en-US,en;q=0.5",
	// 		"Origin":          "https://html.duckduckgo.com",
	// 		"Cookie":          "kl=wt-wt",
	// 	}
	// }
	// ddg, err := duckduckgo.NewTool(ctx, &duckduckgo.Config{
	// 	ToolName:   "search",                                   // 工具名称
	// 	ToolDesc:   "search web for information by duckduckgo", // 工具描述
	// 	Region:     ddgsearch.RegionWT,                         // 搜索地区
	// 	MaxResults: 10,                                         // 每页结果数量
	// 	SafeSearch: ddgsearch.SafeSearchOff,                    // 安全搜索级别
	// 	TimeRange:  ddgsearch.TimeRangeAll,                     // 时间范围
	// 	DDGConfig:  ddgConfig,                                  // DuckDuckGo 配置
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("创建DuckDuckGo工具失败: %w", err)
	// }

	ddg, err := newWebSearchTool(ctx, proxy)
	if err != nil {
		return nil, fmt.Errorf("创建DuckDuckGo工具失败: %w", err)
	}

	// 由于我们不能直接修改 eino 工具的名称，我们需要在生成 toolKey 时使用自定义名称
	// 在 SyncInternalTools 中使用工具信息时会自动使用我们定义的常量

	tools = append(tools, seqThinking, ddg)
	// 这里可以继续添加其他内置工具...

	return tools, nil
}

// GetInternalToolMap 返回内置工具的映射表，键为工具名称，值为工具实例
func GetInternalToolMap(ctx context.Context, proxy string) (map[string]tool.BaseTool, error) {
	tools, err := GetInternalTools(ctx, proxy)
	if err != nil {
		return nil, err
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range tools {
		toolInfo, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取工具 %T 信息失败: %w", t, err)
		}
		toolMap[toolInfo.Name] = t
	}

	return toolMap, nil
}
