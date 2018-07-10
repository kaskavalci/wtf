package markets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"bytes"

	"github.com/senorprogrammer/wtf/wtf"
)

type Widget struct {
	wtf.TextWidget
	result string
	colors struct {
		name, value string
	}
}

type data struct {
	Selling        float64 `json:"selling"`
	UpdateDate     float64 `json:"update_date"`
	Buying         float64 `json:"buying"`
	ChangeRate     float64 `json:"change_rate"`
	Name           string  `json:"name"`
	FullName       string  `json:"full_name"`
	ShortName      string  `json:"short_name"`
	SourceName     string  `json:"source_name"`
	SourceFullName string  `json:"source_full_name"`
	Code           string  `json:"code"`
}

func NewWidget() *Widget {
	widget := Widget{
		TextWidget: wtf.NewTextWidget(" Markets ", "markets", false),
	}

	widget.View.SetWrap(false)

	widget.config()

	return &widget
}

func (widget *Widget) Refresh() {
	if widget.Disabled() {
		return
	}

	widget.UpdateRefreshedAt()
	widget.ipinfo()
	widget.View.Clear()

	widget.View.SetText(widget.result)
}

func (widget *Widget) get(url string) ([]data, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return nil, err
	}
	req.Header.Set("User-Agent", "curl")
	response, err := client.Do(req)
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return nil, err
	}
	defer response.Body.Close()
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return nil, err
	}
	var info []data
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return nil, err
	}

	return info, nil
}

//this method reads the config and calls ipinfo for ip information
func (widget *Widget) ipinfo() {
	goldList, err := widget.get("https://www.doviz.com/api/v1/golds/all/latest")
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return
	}

	fxList, err := widget.get("https://www.doviz.com/api/v1/currencies/all/latest")
	if err != nil {
		widget.result = fmt.Sprintf("%s", err.Error())
		return
	}

	info := append(goldList, fxList...)

	widget.setResult(info)
}

// read module configs
func (widget *Widget) config() {
	nameColor, valueColor := wtf.Config.UString("wtf.mods.markets.colors.name", "red"), wtf.Config.UString("wtf.mods.markets.colors.value", "white")
	widget.colors.name = nameColor
	widget.colors.value = valueColor
}

func (widget *Widget) setResult(info []data) {
	resultTemplate, err := template.New("markets_result").Parse(
		formatableText("Gram Altin ", "gramaltin") +
			formatableText("USD\t\t", "amerikandolari") +
			formatableText("EUR\t\t", "euro") +
			formatableText("GBP\t\t", "sterlin"),
	)
	if err != nil {
		widget.result = fmt.Sprintf("cannot create template %s", err.Error())
		return
	}

	resultBuffer := new(bytes.Buffer)

	infoMap := map[string]string{
		"nameColor":  widget.colors.name,
		"valueColor": widget.colors.value,
	}
	for _, item := range info {
		key := strings.Replace(item.Name, "-", "", -1)
		infoMap[key] = strconv.FormatFloat(item.Selling, 'f', 3, 32) + " ₺"
	}

	resultTemplate.Execute(resultBuffer, infoMap)

	widget.result = resultBuffer.String()
}

func formatableText(key, value string) string {
	return fmt.Sprintf(" [{{.nameColor}}]%s: [{{.valueColor}}]{{.%s}}\n", key, value)
}
