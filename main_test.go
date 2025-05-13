package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestCreateAgent(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		coreTraits  string
		expectError bool
	}{
		{
			name:        "正常创建",
			agentName:   "测试智能体",
			coreTraits:  "逻辑思维,创新能力",
			expectError: false,
		},
		{
			name:        "空名称",
			agentName:   "",
			coreTraits:  "逻辑思维",
			expectError: true,
		},
		{
			name:        "空特征",
			agentName:   "测试智能体",
			coreTraits:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			request := mcp.CallToolRequest{
				Params: ParamsStruct{
					Arguments: map[string]interface{}{
						"agent_name":  tt.agentName,
						"core_traits": tt.coreTraits,
					},
				},
			}

			result, err := createToolHandler(ctx, request)
			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但获得成功")
				}
				return
			}

			if err != nil {
				t.Errorf("未期望错误但发生错误: %v", err)
				return
			}

			if result == nil {
				t.Error("期望非空结果但获得空结果")
				return
			}
		})
	}
}

func TestGetAgent(t *testing.T) {
	// 创建测试用智能体
	testAgent := &Agent{
		ID:          uuid.New().String(),
		Name:        "测试智能体",
		CoreTraits:  "测试特征",
		Personality: "测试性格",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	agents[testAgent.ID] = testAgent

	tests := []struct {
		name        string
		agentID     string
		expectError bool
	}{
		{
			name:        "获取存在的智能体",
			agentID:     testAgent.ID,
			expectError: false,
		},
		{
			name:        "获取不存在的智能体",
			agentID:     "non-existent-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			request := mcp.CallToolRequest{
				Params: ParamsStruct{
					Arguments: map[string]interface{}{
						"agent_id": tt.agentID,
					},
				},
			}

			result, err := getAgentHandler(ctx, request)
			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但获得成功")
				}
				return
			}

			if err != nil {
				t.Errorf("未期望错误但发生错误: %v", err)
				return
			}

			if result == nil {
				t.Error("期望非空结果但获得空结果")
				return
			}
		})
	}
}

func TestListAgents(t *testing.T) {
	// 清空智能体列表
	agents = make(map[string]*Agent)

	// 添加测试用智能体
	testAgents := []*Agent{
		{
			ID:          uuid.New().String(),
			Name:        "测试智能体1",
			CoreTraits:  "特征1",
			Personality: "性格1",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          uuid.New().String(),
			Name:        "测试智能体2",
			CoreTraits:  "特征2",
			Personality: "性格2",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	for _, agent := range testAgents {
		agents[agent.ID] = agent
	}

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: ParamsStruct{},
	}

	result, err := listAgentsHandler(ctx, request)
	if err != nil {
		t.Errorf("列出智能体时发生错误: %v", err)
		return
	}

	if result == nil {
		t.Error("期望非空结果但获得空结果")
		return
	}
}

func TestDeleteAgent(t *testing.T) {
	// 创建测试用智能体
	testAgent := &Agent{
		ID:          uuid.New().String(),
		Name:        "测试智能体",
		CoreTraits:  "测试特征",
		Personality: "测试性格",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	agents[testAgent.ID] = testAgent

	tests := []struct {
		name        string
		agentID     string
		expectError bool
	}{
		{
			name:        "删除存在的智能体",
			agentID:     testAgent.ID,
			expectError: false,
		},
		{
			name:        "删除不存在的智能体",
			agentID:     "non-existent-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			request := mcp.CallToolRequest{
				Params: ParamsStruct{
					Arguments: map[string]interface{}{
						"agent_id": tt.agentID,
					},
				},
			}

			result, err := deleteAgentHandler(ctx, request)
			if tt.expectError {
				if err == nil {
					t.Errorf("期望错误但获得成功")
				}
				return
			}

			if err != nil {
				t.Errorf("未期望错误但发生错误: %v", err)
				return
			}

			if result == nil {
				t.Error("期望非空结果但获得空结果")
				return
			}

			// 验证智能体是否已被删除
			if _, exists := agents[tt.agentID]; exists {
				t.Error("智能体应该被删除但仍然存在")
			}
		})
	}
}
