package handler

import (
	"encoding/json"
	"fmt"
	"ncobase/plugin/ai/service"
	"ncobase/plugin/ai/structs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

type AIHandlerInterface interface {
	Status(c *gin.Context)
	Providers(c *gin.Context)
	Health(c *gin.Context)
	Models(c *gin.Context)
	Complete(c *gin.Context)
	Stream(c *gin.Context)
	Embed(c *gin.Context)
}

type aiHandler struct {
	svc service.AIServiceInterface
}

func NewAIHandler(svc service.AIServiceInterface) AIHandlerInterface {
	return &aiHandler{svc: svc}
}

// Status returns safe AI runtime status without exposing secrets.
func (h *aiHandler) Status(c *gin.Context) {
	result, err := h.svc.Status(c.Request.Context())
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *aiHandler) Providers(c *gin.Context) {
	result, err := h.svc.Providers(c.Request.Context())
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *aiHandler) Health(c *gin.Context) {
	result, err := h.svc.Health(c.Request.Context())
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *aiHandler) Models(c *gin.Context) {
	result, err := h.svc.Models(c.Request.Context(), c.Query("provider"))
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *aiHandler) Complete(c *gin.Context) {
	var input structs.CompleteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body", err.Error()))
		return
	}
	result, err := h.svc.Complete(c.Request.Context(), &input)
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *aiHandler) Stream(c *gin.Context) {
	var input structs.CompleteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body", err.Error()))
		return
	}

	session, err := h.svc.StartStream(c.Request.Context(), &input)
	if err != nil {
		handleAIError(c, err)
		return
	}

	w := c.Writer
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	writeSSE(w, "start", structs.StreamStartResponse{RunID: session.Run.ID, OperationID: session.Run.OperationID})
	flushSSE(w)

	var content strings.Builder
	var final *structs.StreamChunkResponse
	var streamErr error

	for chunk := range session.Chunks {
		response := structs.StreamChunkResponse{
			RunID:           session.Run.ID,
			Content:         chunk.Content,
			Reasoning:       chunk.Reasoning,
			Done:            chunk.Done,
			FinishReason:    chunk.FinishReason,
			InputTokens:     chunk.InputTokens,
			OutputTokens:    chunk.OutputTokens,
			TotalTokens:     chunk.TokensUsed,
			ReasoningTokens: chunk.ReasoningTokens,
			CacheUsage: structs.CacheUsage{
				CreatedTokens: chunk.CacheUsage.CreatedTokens,
				ReadTokens:    chunk.CacheUsage.ReadTokens,
			},
		}
		if chunk.Error != nil {
			streamErr = chunk.Error
			response.Error = chunk.Error.Error()
			writeSSE(w, "error", response)
			flushSSE(w)
			break
		}
		if chunk.Content != "" {
			content.WriteString(chunk.Content)
		}
		if chunk.Done {
			final = &response
		}
		writeSSE(w, "chunk", response)
		flushSSE(w)
	}

	run, finishErr := h.svc.FinishStream(c.Request.Context(), session, final, content, streamErr)
	if finishErr != nil {
		writeSSE(w, "error", map[string]any{"run_id": session.Run.ID, "error": finishErr.Error()})
		flushSSE(w)
		return
	}
	writeSSE(w, "run", run)
	flushSSE(w)
}

func (h *aiHandler) Embed(c *gin.Context) {
	var input structs.EmbedRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body", err.Error()))
		return
	}
	result, err := h.svc.Embed(c.Request.Context(), &input)
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func writeSSE(w gin.ResponseWriter, event string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte(`{"error":"failed to encode stream payload"}`)
	}
	_, _ = fmt.Fprintf(w, "event: %s\n", event)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", raw)
}

func flushSSE(w gin.ResponseWriter) {
	if flusher, ok := any(w).(http.Flusher); ok {
		flusher.Flush()
	}
}
