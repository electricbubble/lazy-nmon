package tool

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/chenjiandongx/go-echarts/charts"
	"github.com/shopspring/decimal"
)

// WorkPath 当前工作路径
var WorkPath *string

// NmonNameFormat nmon结果文件命名格式
var NmonNameFormat *string

// DirReport 保存解析后页面的文件夹
const DirReport string = "report"

// SumOfFloat 计算float的和
func SumOfFloat(value ...float64) float64 {
	sum := decimal.NewFromFloat32(0)
	for _, v := range value {
		sum = sum.Add(decimal.NewFromFloat(v))
	}
	ret, _ := sum.Float64()
	return ret
}

// GetFloatFromDecimal decimal.Decimal转float64
func GetFloatFromDecimal(value decimal.Decimal) float64 {
	ret, _ := value.Float64()
	return ret
}

// GetFloatFromString 字符串转float64
func GetFloatFromString(value string) float64 {
	n, _ := decimal.NewFromString(value)
	ret, _ := n.Float64()
	return ret
}

// SumOfSpecifiedColumns 返回当前行的指定列之和
func SumOfSpecifiedColumns(line string, columns []int) float64 {
	arr := strings.Split(line, ",")
	sum := decimal.NewFromFloat32(0)
	for _, index := range columns {
		n, _ := decimal.NewFromString(arr[index])
		sum = sum.Add(n)
	}
	ret, _ := sum.Float64()
	return ret
}

// SumOfEachColumns 返回当前行的列之和(不包含前两列)
func SumOfEachColumns(line string) float64 {
	arr := strings.Split(line, ",")
	sum := decimal.NewFromFloat32(0)
	for i := 2; i < len(arr); i++ {
		n, err := decimal.NewFromString(arr[i])
		if err != nil {
			fmt.Println(err, "该值将当作0完成后续计算")
			n = decimal.NewFromFloat32(0)
		}
		sum = sum.Add(n)
	}
	ret, _ := sum.Float64()
	return ret
}

// ParseDate 转换目标时间格式为 yyyymmdd_hhmmss
func ParseDate(date string) string {
	var format = "20060102_150405"
	t, err := dateparse.ParseAny(date)
	if err != nil {
		return fmt.Sprintf("%s_%v", err.Error(), time.Now().Format(format))
	}
	return fmt.Sprintf("%v", t.Format(format))
}

// GetNmonFileName 获取一个nmon结果文件名
// 以当前工作路径为根路径
func GetNmonFileName() (string, error) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s", *WorkPath))
	if err != nil {
		return "", err
	}
	for _, info := range files {
		f := info.Name()
		if !info.IsDir() && strings.Index(f, ".") != -1 && f[len(f)-5:] == ".nmon" {
			return f, nil
		}
	}
	return "", errors.New("无.nmon结果文件")
}

// MkdirIfNotExist 如果指定文件夹不存在则创建
// 以当前工作路径为根路径
func MkdirIfNotExist(destDir string) {
	destDir = path.Join(*WorkPath, destDir)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		// 文件夹不存在则创建
		os.MkdirAll(destDir, os.ModePerm)
	}
}

// MoveFile 移动文件
// 以当前工作路径为根路径
func MoveFile(file, destDir string) error {
	destDir = path.Join(*WorkPath, destDir, file)
	file = path.Join(*WorkPath, file)
	err := os.Rename(file, destDir)
	if err != nil {
		return err
	}
	return nil
}

// GetCurrentPath 获取当前工作路径
func GetCurrentPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

// CreateDisplayPage 创建一个同时显示4张图表📈的HTML
// nmonName 以nmon结果文件名作为文件夹保存所有图表
func CreateDisplayPage(nmonName string) error {
	pathPage := path.Join(*WorkPath, DirReport, nmonName, "index.html")
	if _, err := os.Stat(pathPage); os.IsNotExist(err) {
		// 文件不存在则创建
		f, err := os.Create(pathPage)
		if err != nil {
			return err
		}
		_, err = io.WriteString(
			f, fmt.Sprintf(templetHTML, nmonName),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveChartAsHTML 保存为HTML文件
// nmonName 以nmon结果文件名作为文件夹保存所有图表
// htmlName 文件名无需扩展名
func SaveChartAsHTML(myChart *charts.Line, nmonName, htmlName string) error {
	MkdirIfNotExist(path.Join(DirReport, nmonName))
	f, err := os.Create(path.Join(*WorkPath, DirReport, nmonName, htmlName+".html"))
	if err != nil {
		return err
	}
	myChart.Render(f)
	return nil
}

// AddXAxis 增加X轴数据
func AddXAxis(myChart *charts.Line, xTime []string, xv ...interface{}) {
	myChart.AddXAxis(xTime)
	for i := 0; i < len(xv); i += 2 {
		if i == len(xv)-2 {
			myChart.AddYAxis(
				xv[i].(string), xv[i+1].([]float64),
				charts.LineStyleOpts{Width: 1.0},
				charts.AreaStyleOpts{Opacity: 0.5},
				// 显示图形上的文本标签
				charts.LabelTextOpts{Show: true},
			)
		} else {
			myChart.AddYAxis(
				xv[i].(string), xv[i+1].([]float64),
				charts.LineStyleOpts{Width: 1.0},
				charts.AreaStyleOpts{Opacity: 0.5},
			)
		}
	}
}

// GenerateGlobalOptions 生成全局设置
// 可以设置成特殊值 'dataMax'，此时取数据在该轴上的最小值作为最小刻度，数值轴有效
func GenerateGlobalOptions(myChart *charts.Line, titleName string, dataMax interface{}) *charts.RectChart {
	return myChart.SetGlobalOptions(
		charts.TitleOpts{
			Title: titleName,
			// Subtitle: fmt.Sprintf("Max: %.1f%%\nAverage: %.1f%%\nMin: %.1f%%", maxMemUsage, averageMemUsage, minMemUsage),
		},
		// 显示工具箱
		charts.ToolboxOpts{Show: true},
		charts.InitOpts{
			// 修改为本地引用
			AssetsHost: "http://127.0.0.1:6060/assets/",
			// 修改html标题
			PageTitle: "lazy nmon",
			Width:     "540px",
			Height:    "300px",
			// 设置主题
			// Theme: "chalk",
		},
		charts.YAxisOpts{
			// 显示分割线
			SplitLine: charts.SplitLineOpts{Show: true},
			// Y轴最大值
			Max: dataMax,
		},
		charts.DataZoomOpts{XAxisIndex: []int{0}, Start: 0, End: 100},
	)
}

var templetHTML = `<!DOCTYPE html>
<html>

<head>
	<meta charset="utf-8">
	<title>%s</title>
</head>

<body>
	<iframe name="CPU_ALL" style="width:49%%;height:400px;" frameborder="0" src="./CPU_ALL.html"></iframe>
	<iframe name="Memory" style="width:49%%;height:400px;" frameborder="0" src="./Memory.html"></iframe>
	<iframe name="Net" style="width:49%%;height:400px;" frameborder="0" src="./Net.html"></iframe>
	<iframe name="Disk" style="width:49%%;height:400px;" frameborder="0" src="./Disk.html"></iframe>
</body>

</html>`
