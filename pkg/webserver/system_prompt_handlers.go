package webserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LubyRuffy/mcpagent/pkg/models"
	"github.com/gorilla/mux"
)

// 系统提示词API请求结构体
type CreateSystemPromptRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Content      string   `json:"content"`
	Placeholders []string `json:"placeholders"`
	IsDefault    bool     `json:"is_default"`
}

// handleListSystemPrompts 列出所有系统提示词配置
func (s *Server) handleListSystemPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := s.systemPromptService.ListPrompts()
	if err != nil {
		http.Error(w, "获取系统提示词配置列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// handleCreateSystemPrompt 创建新的系统提示词配置
func (s *Server) handleCreateSystemPrompt(w http.ResponseWriter, r *http.Request) {
	var req CreateSystemPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析请求失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 创建新的系统提示词模型
	prompt := &models.SystemPromptModel{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		IsDefault:   req.IsDefault,
		IsActive:    true,
	}

	// 设置占位符
	if err := prompt.SetPlaceholdersFromStringSlice(req.Placeholders); err != nil {
		http.Error(w, "设置占位符失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 保存到数据库
	if err := s.systemPromptService.CreatePrompt(prompt); err != nil {
		http.Error(w, "创建系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prompt)
}

// handleGetSystemPrompt 获取特定的系统提示词配置
func (s *Server) handleGetSystemPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	prompt, err := s.systemPromptService.GetPrompt(uint(id))
	if err != nil {
		if err == models.ErrSystemPromptNotFound {
			http.Error(w, "系统提示词配置不存在", http.StatusNotFound)
		} else {
			http.Error(w, "获取系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

// handleUpdateSystemPrompt 更新系统提示词配置
func (s *Server) handleUpdateSystemPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var req CreateSystemPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析请求失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 创建更新的系统提示词模型
	updates := &models.SystemPromptModel{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		IsDefault:   req.IsDefault,
	}

	// 设置占位符
	if err := updates.SetPlaceholdersFromStringSlice(req.Placeholders); err != nil {
		http.Error(w, "设置占位符失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新数据库
	if err := s.systemPromptService.UpdatePrompt(uint(id), updates); err != nil {
		if err == models.ErrSystemPromptNotFound {
			http.Error(w, "系统提示词配置不存在", http.StatusNotFound)
		} else if err == models.ErrSystemPromptNameExists {
			http.Error(w, "系统提示词名称已存在", http.StatusConflict)
		} else {
			http.Error(w, "更新系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 获取更新后的配置
	prompt, err := s.systemPromptService.GetPrompt(uint(id))
	if err != nil {
		http.Error(w, "获取更新后的系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

// handleDeleteSystemPrompt 删除系统提示词配置
func (s *Server) handleDeleteSystemPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	if err := s.systemPromptService.DeletePrompt(uint(id)); err != nil {
		if err == models.ErrSystemPromptNotFound {
			http.Error(w, "系统提示词配置不存在", http.StatusNotFound)
		} else {
			http.Error(w, "删除系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleSetDefaultSystemPrompt 设置默认系统提示词配置
func (s *Server) handleSetDefaultSystemPrompt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	if err := s.systemPromptService.SetDefaultPrompt(uint(id)); err != nil {
		if err == models.ErrSystemPromptNotFound {
			http.Error(w, "系统提示词配置不存在", http.StatusNotFound)
		} else {
			http.Error(w, "设置默认系统提示词配置失败: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
