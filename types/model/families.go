package model

// SingleParallelArchitectures contains model families that are not safe with num_parallel > 1
var SingleParallelArchitectures = []string{
	"mllama",
	"qwen3vl",
	"qwen3vlmoe",
	"qwen35",
	"qwen35moe",
	"qwen3next",
	"lfm2",
	"lfm2moe",
	"nemotron_h",
	"nemotron_h_moe",
	"nemotron_h_omni",
}
