// 提供常用的 html/template 函数
package funcmap

import (
	"html/template"
)

func rawHTML(text string) template.HTML {
	return template.HTML(text)
}

func rawHTMLAttr(text string) template.HTMLAttr {
	return template.HTMLAttr(text)
}

func rawCSS(text string) template.CSS {
	return template.CSS(text)
}

func rawJS(text string) template.JS {
	return template.JS(text)
}

func rawJSStr(text string) template.JSStr {
	return template.JSStr(text)
}

func rawURL(text string) template.URL {
	return template.URL(text)
}

// 提供常用的 html/template 函数
// RawFuncMap 为 html/template.FuncMap 类型
// 里面提供了 rawHTML, rawHTMLAttr, rawCSS, rawJS, rawJSStr, rawURL 函数,
// 这些函数可以在模板中应用, 如 {{rawHTML `<br>`}}
var RawFuncMap = template.FuncMap{
	"rawHTML":     rawHTML,
	"rawHTMLAttr": rawHTMLAttr,
	"rawCSS":      rawCSS,
	"rawJS":       rawJS,
	"rawJSStr":    rawJSStr,
	"rawURL":      rawURL,
}
