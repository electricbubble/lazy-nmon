package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"lazy_nmon/tool"
	"log"
	"os"
	"path"
	"strings"

	"github.com/chenjiandongx/go-echarts/charts"
	"github.com/shopspring/decimal"
)

const (
	dirProcessed = "processed"
	dirResult    = "result"
)

func main() {
	// 默认工作路径为lazy_nmon的路径
	tool.WorkPath = flag.String("wp", tool.GetCurrentPath(), "")
	//tool.WorkPath = flag.String("wp", tool.GetCurrentPath(), "指定当前工作路径")
	// 默认命名格式为 ${name}_${users}u_${duration}m_${now}
	tool.NmonNameFormat = flag.String("fmt", "${name}_${users}u_${duration}m_${now}", "/home/ec2-user/test/goproj/nmon_files/ip-172-31-10-38_200717_0733.nmon")
	flag.Parse()

	// fmt.Printf("nmon结果文件命名格式: %s\n", *tool.NmonNameFormat)

	// fmt.Println("时间转换为：", tool.ParseDate("16-MAR-2019 00:02:47"))

	tool.MkdirIfNotExist(dirProcessed)
	// tool.MkdirIfNotExist(dirResult)
	// tool.MkdirIfNotExist(tool.DirReport)
	fileName, err := tool.GetNmonFileName()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("获取nmon结果文件：", fileName)

	file, err := os.Open(path.Join(*tool.WorkPath, fileName))
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	hasZZZZ := false
	hasCPUAll := false
	hasDiskRead := false
	hasDiskWrite := false
	hasMem := false
	hasNet := false
	indexNetRead := make([]int, 0)
	indexNetWrite := make([]int, 0)
	sliceZZZZTime := make([]string, 0, 1024)
	maxCPUUsage, averageCPUUsage, minCPUUsage := 0.0, 0.0, 100.0
	sliceCPUUser := make([]float64, 0, 1024)
	sliceCPUSys := make([]float64, 0, 1024)
	sliceCPUWait := make([]float64, 0, 1024)
	sliceCPUUsage := make([]float64, 0, 1024)
	sliceDiskRead := make([]float64, 0, 1024)
	sliceDiskWrite := make([]float64, 0, 1024)
	maxMemUsage, averageMemUsage, minMemUsage, memoryTotal := 0.0, 0.0, 100.0, 0.0
	sliceMemFree := make([]float64, 0, 1024)
	sliceMemCached := make([]float64, 0, 1024)
	sliceMemBuffers := make([]float64, 0, 1024)
	sliceMemUsage := make([]float64, 0, 1024)
	sliceNetReadTotal := make([]float64, 0, 1024)
	sliceNetWriteTotal := make([]float64, 0, 1024)
	for {
		line, isPrefix, err := reader.ReadLine()
		// 解决单行字节数大于4096的情况
		for isPrefix && err == nil {
			var bs []byte
			bs, isPrefix, err = reader.ReadLine()
			line = append(line, bs...)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			return
		}
		strLine := string(line)
		arr := strings.Split(strLine, ",")
		if !hasZZZZ && strings.HasPrefix(strLine, "ZZZZ,") {
			hasZZZZ = true
			// ZZZZ没有标题栏，所以不需要continue
		}
		if !hasCPUAll && strings.HasPrefix(strLine, "CPU_ALL,CPU Total") {
			hasCPUAll = true
			continue
		}
		if !hasDiskRead && strings.HasPrefix(strLine, "DISKREAD,Disk Read KB/s") {
			hasDiskRead = true
			continue
		}
		if !hasDiskWrite && strings.HasPrefix(strLine, "DISKWRITE,Disk Write KB/s") {
			hasDiskWrite = true
			continue
		}
		if !hasMem && strings.HasPrefix(strLine, "MEM,Memory MB") {
			hasMem = true
			continue
		}
		if !hasNet && strings.HasPrefix(strLine, "NET,Network I/O") {
			hasNet = true
			// 或许存在多个网络适配器
			for i, columnName := range arr {
				if strings.HasSuffix(columnName, "-read-KB/s") {
					indexNetRead = append(indexNetRead, i)
				}
				if strings.HasSuffix(columnName, "-write-KB/s") {
					indexNetWrite = append(indexNetWrite, i)
				}
			}
			continue
		}
		if hasZZZZ && strings.HasPrefix(strLine, "ZZZZ,") { //&& iz <= 5 {
			sliceZZZZTime = append(sliceZZZZTime, arr[2])
			// fmt.Println(strLine, tool.ParseDate(arr[3]+" "+arr[2]))
			// iz++
			continue
		}
		if hasCPUAll && strings.HasPrefix(strLine, "CPU_ALL,") { //&& ic <= 20 {
			uu := tool.GetFloatFromString(arr[2])
			su := tool.GetFloatFromString(arr[3])
			wu := tool.GetFloatFromString(arr[4])
			use := tool.SumOfFloat(uu, su)
			sliceCPUUser = append(sliceCPUUser, uu)
			sliceCPUSys = append(sliceCPUSys, su)
			sliceCPUWait = append(sliceCPUWait, wu)
			sliceCPUUsage = append(sliceCPUUsage, use)
			if maxCPUUsage < use {
				maxCPUUsage = use
			}
			if minCPUUsage > use {
				minCPUUsage = use
			}
			averageCPUUsage = tool.SumOfFloat(averageCPUUsage, use)
			// fmt.Println(strLine, "\tUser%", arr[2], "\tSys%", arr[3], "\tWait%", arr[4])
			// ic++
			continue
		}
		if hasDiskRead && strings.HasPrefix(strLine, "DISKREAD,") { //&& id <= 20 {
			sliceDiskRead = append(sliceDiskRead, tool.SumOfEachColumns(strLine))
			// fmt.Println(strLine, "\tDisk Read KB/s", tool.SumOfEachColumns(strLine))
			// id++
			continue
		}
		if hasDiskWrite && strings.HasPrefix(strLine, "DISKWRITE,") { //&& iw <= 5 {
			sliceDiskWrite = append(sliceDiskWrite, tool.SumOfEachColumns(strLine))
			// fmt.Println(strLine, "\tDisk Write KB/s", tool.SumOfEachColumns(strLine))
			// iw++
			continue
		}
		if hasMem && strings.HasPrefix(strLine, "MEM,") { //&& im <= 5 {
			mTotal, _ := decimal.NewFromString(arr[2])
			mFree, _ := decimal.NewFromString(arr[6])
			mCached, _ := decimal.NewFromString(arr[11])
			mBuffers, _ := decimal.NewFromString(arr[14])
			mUsage := mTotal.Sub(mFree).Sub(mCached).Sub(mBuffers).DivRound(mTotal, 4).Mul(decimal.NewFromFloat32(100))
			nu, _ := mUsage.Float64()
			if memoryTotal == 0.0 {
				memoryTotal = tool.GetFloatFromDecimal(mTotal)
			}
			if maxMemUsage < nu {
				maxMemUsage = nu
			}
			if minMemUsage > nu {
				minMemUsage = nu
			}
			averageMemUsage = tool.SumOfFloat(averageMemUsage, nu)
			sliceMemFree = append(sliceMemFree, tool.GetFloatFromDecimal(mFree))
			sliceMemCached = append(sliceMemCached, tool.GetFloatFromDecimal(mCached))
			sliceMemBuffers = append(sliceMemBuffers, tool.GetFloatFromDecimal(mBuffers))
			sliceMemUsage = append(sliceMemUsage, nu)
			// // fmt.Println(strLine)
			// fmt.Println("memtotal:", arr[2], "\tmemfree:", arr[6], "\tcached:", arr[11], "\tbuffers:", arr[14], "\tmemuse%:", n)
			// im++
			continue
		}
		if hasNet && strings.HasPrefix(strLine, "NET,") { //&& in <= 20 {
			sliceNetReadTotal = append(sliceNetReadTotal, tool.SumOfSpecifiedColumns(strLine, indexNetRead))
			sliceNetWriteTotal = append(sliceNetWriteTotal, tool.SumOfSpecifiedColumns(strLine, indexNetWrite)*-1)
			// fmt.Println(strLine, "\tTotal-Read:", tool.SumOfSpecifiedColumns(strLine, indexNetRead), "\tTotal-Write:", tool.SumOfSpecifiedColumns(strLine, indexNetWrite))
			// in++
			continue
		}

	}
	if !hasZZZZ {
		log.Println("解析nmon结果文件失败")
		return
	}

	fileNameWithoutExt := fileName[:strings.LastIndex(fileName, ".")]

	if hasCPUAll {
		averageCPUUsage = averageCPUUsage / float64(len(sliceCPUUsage))
		cpuChart := charts.NewLine()
		tool.GenerateGlobalOptions(cpuChart, "CPU_ALL", 100)
		tool.AddXAxis(cpuChart, sliceZZZZTime, "User%", sliceCPUUser, "Sys%", sliceCPUSys, "Wait%", sliceCPUWait, "Use%", sliceCPUUsage)
		tool.SaveChartAsHTML(cpuChart, fileNameWithoutExt, "CPU_ALL")
	}

	if hasMem {
		averageMemUsage = averageMemUsage / float64(len(sliceMemUsage))
		memChart := charts.NewLine()
		tool.GenerateGlobalOptions(memChart, "Memory", 100)
		tool.AddXAxis(memChart, sliceZZZZTime, "Use%", sliceMemUsage)
		tool.SaveChartAsHTML(memChart, fileNameWithoutExt, "Memory")
	}

	if hasNet {
		netChart := charts.NewLine()
		tool.GenerateGlobalOptions(netChart, "Net (KB/s) ", "dataMax")
		tool.AddXAxis(netChart, sliceZZZZTime, "Total-Read", sliceNetReadTotal, "Total-Write(-ve)", sliceNetWriteTotal)
		tool.SaveChartAsHTML(netChart, fileNameWithoutExt, "Net")
	}

	if hasDiskRead && hasDiskWrite {
		diskChart := charts.NewLine()
		tool.GenerateGlobalOptions(diskChart, "Disk (KB/s) ", "dataMax")
		tool.AddXAxis(diskChart, sliceZZZZTime, "Disk-Read", sliceDiskRead, "Disk-Write", sliceDiskWrite)
		tool.SaveChartAsHTML(diskChart, fileNameWithoutExt, "Disk")
	}

	err = tool.CreateDisplayPage(fileNameWithoutExt)
	if err != nil {
		log.Println("index.html", err)
	}

	// TODO 待完成解析后再移动文件
	// tool.MoveFile(fileName, dirProcessed)

}
