package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	proc "github.com/shirou/gopsutil/v3/process"
	"log"
	"sort"
)

type ConsolidatedProcess struct {
	p             *proc.Process
	CPUPercent    float64
	MemoryPercent float32
	Command       string
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/metrics", getMetrics)

	log.Fatal(app.Listen(":9101"))
}

func getMetrics(c *fiber.Ctx) error {
	pIds, _ := proc.Processes()
	processes := make([]ConsolidatedProcess, len(pIds))
	for _, pId := range pIds {
		cpuPercent, _ := pId.CPUPercent()
		memPercent, _ := pId.MemoryPercent()
		cmd, _ := pId.Cmdline()
		processes = append(processes, ConsolidatedProcess{
			p:             pId,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
			Command:       cmd,
		})
	}

	sort.SliceStable(processes, func(i, j int) bool {
		return processes[i].MemoryPercent > processes[j].MemoryPercent
	})

	metricString := "#HELP top_processes_by_memory mem utilization by top 5 processes\n#TYPE top_processes_by_memory gauge\n"
	for _, i := range processes[:5] {
		metricString += fmt.Sprintf("top_processes_by_memory{app=\"%s\", pid=\"%v\"} %f\n", i.Command, i.p.Pid ,i.MemoryPercent)
	}

	return c.SendString(metricString)

}
