# lazy-nmon
nmon辅助工具，可直接提取nmon的结果文件中的CPU、Memory、Net、Disk生成图表展示。  
**仅为试验性尝试，不打算再调整。**

## 代码运行须知
`main.go`存放在`cmd/app`中，所以在调试阶段务必添加`wp`指定当前项目所在的绝对路径，以下为Visual Studio Code的`launch.json`
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "golang",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "env": {},
            "args": [
                "-wp","/Users/hero/Documents/Workspace/Go/src/lazy_nmon"
            ]
        }
    ]
}
```

## 感谢
参考[eazyNmon](https://github.com/mzky/easyNmon)项目  
图表为[go-echarts](https://github.com/chenjiandongx/go-echarts)提供
