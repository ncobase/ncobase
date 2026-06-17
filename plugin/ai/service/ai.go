package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"ncobase/plugin/ai/data/repository"
	"ncobase/plugin/ai/structs"
	"sort"
	"strings"
	"time"

	"github.com/ncobase/deebus"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/utils/nanoid"
)

type AIServiceInterface interface {
	Status(ctx context.Context) (*structs.StatusResponse, error)
	Providers(ctx context.Context) ([]structs.ProviderStatus, error)
	Health(ctx context.Context) ([]structs.ProviderHealth, error)
	Models(ctx context.Context, provider string) (map[string][]string, error)
	Actions(ctx context.Context) ([]structs.ActionDescriptor, error)
	Complete(ctx context.Context, input *structs.CompleteRequest) (*structs.CompleteResponse, error)
	StartStream(ctx context.Context, input *structs.CompleteRequest) (*StreamSession, error)
	FinishStream(ctx context.Context, session *StreamSession, final *structs.StreamChunkResponse, content strings.Builder, err error) (*structs.Run, error)
	Embed(ctx context.Context, input *structs.EmbedRequest) (*structs.EmbedResponse, error)
	RunAction(ctx context.Context, action string, input *structs.ActionRequest) (*structs.ActionResponse, error)
}

type aiService struct {
	runRepo repository.RunRepositoryInterface
	manager *clientManager
}

type StreamSession struct {
	Run         *structs.Run
	Chunks      <-chan *deebus.StreamChunk
	Config      *structs.Config
	Cancel      context.CancelFunc
	StartedAt   time.Time
	Model       string
	Provider    string
	RequestHash string
}

func NewAIService(runRepo repository.RunRepositoryInterface, configProvider ConfigProvider) AIServiceInterface {
	return &aiService{
		runRepo: runRepo,
		manager: newClientManager(configProvider),
	}
}

func (s *aiService) Status(ctx context.Context) (*structs.StatusResponse, error) {
	status, _, err := s.manager.status(ctx)
	return status, err
}

func (s *aiService) Providers(ctx context.Context) ([]structs.ProviderStatus, error) {
	status, _, err := s.manager.status(ctx)
	if err != nil {
		return nil, err
	}
	return status.Providers, nil
}

func (s *aiService) Health(ctx context.Context) ([]structs.ProviderHealth, error) {
	status, cfg, err := s.manager.status(ctx)
	if err != nil {
		return nil, err
	}
	if !status.Ready {
		return nil, ErrAIUnconfigured
	}
	client, err := s.manager.get(ctx, cfg, "", nil)
	if err != nil {
		return nil, err
	}

	checkCtx, cancel := withPolicyTimeout(ctx, cfg)
	defer cancel()

	results := client.Health(checkCtx)
	checkedAt := time.Now().UnixMilli()
	providerTypes := make(map[string]string, len(cfg.Providers))
	for _, provider := range cfg.Providers {
		providerTypes[provider.Name] = provider.Type
	}

	health := make([]structs.ProviderHealth, 0, len(results))
	for name, resultErr := range results {
		item := structs.ProviderHealth{
			Name:      name,
			Type:      providerTypes[name],
			Healthy:   resultErr == nil,
			CheckedAt: checkedAt,
		}
		if resultErr != nil {
			_, item.Error = sanitizeError(resultErr, cfg.Safety.MaxErrorChars)
		}
		health = append(health, item)
	}
	return health, nil
}

func (s *aiService) Models(ctx context.Context, provider string) (map[string][]string, error) {
	status, cfg, err := s.manager.status(ctx)
	if err != nil {
		return nil, err
	}
	if !status.Configured {
		return nil, ErrAIUnconfigured
	}
	client, err := s.manager.get(ctx, cfg, "", nil)
	if err != nil {
		return nil, err
	}

	providers := make([]string, 0, len(status.Providers))
	if provider != "" {
		providers = append(providers, provider)
	} else {
		for _, item := range status.Providers {
			if item.Enabled && item.Configured {
				providers = append(providers, item.Name)
			}
		}
	}

	checkCtx, cancel := withPolicyTimeout(ctx, cfg)
	defer cancel()

	models := make(map[string][]string, len(providers))
	for _, providerName := range providers {
		values, err := client.ListModels(checkCtx, providerName)
		if err != nil {
			return nil, err
		}
		models[providerName] = values
	}
	return models, nil
}

func (s *aiService) Actions(ctx context.Context) ([]structs.ActionDescriptor, error) {
	cfg, err := s.manager.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	allowed := stringSet(cfg.Policy.AllowedActions)
	actions := make([]structs.ActionDescriptor, 0, len(actionRegistry))
	for key, descriptor := range actionRegistry {
		if _, ok := allowed[key]; ok {
			actions = append(actions, descriptor)
		}
	}
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Domain == actions[j].Domain {
			return actions[i].Key < actions[j].Key
		}
		return actions[i].Domain < actions[j].Domain
	})
	return actions, nil
}

func (s *aiService) Complete(ctx context.Context, input *structs.CompleteRequest) (*structs.CompleteResponse, error) {
	return s.complete(ctx, input, structs.RunModeComplete)
}

func (s *aiService) complete(ctx context.Context, input *structs.CompleteRequest, mode structs.RunMode) (*structs.CompleteResponse, error) {
	cfg, err := s.manager.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	req, metadata, requestHash, model, provider, err := s.prepareCompletion(ctx, cfg, input, mode)
	if err != nil {
		return nil, err
	}

	client, err := s.manager.get(ctx, cfg, model, nil)
	if err != nil {
		return nil, err
	}

	run, err := s.createRun(ctx, mode, input.Action, provider, model, requestHash, metadata)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := withPolicyTimeout(ctx, cfg)
	defer cancel()

	started := time.Now()
	resp, err := client.Complete(callCtx, req)
	if err != nil {
		_, _ = s.failRun(ctx, cfg, run.ID, started, err)
		return nil, err
	}

	updated, err := s.succeedRun(ctx, cfg, run.ID, started, resp.Provider, resp.Model, cfg.Model.Primary, responseHash(resp.Content), resp.InputTokens, resp.OutputTokens, resp.TokensUsed, resp.ReasoningTokens, resp.CacheUsage.CreatedTokens, resp.CacheUsage.ReadTokens)
	if err != nil {
		return nil, err
	}

	return completionResponse(updated.ID, updated.OperationID, resp, cfg), nil
}

func (s *aiService) StartStream(ctx context.Context, input *structs.CompleteRequest) (*StreamSession, error) {
	cfg, err := s.manager.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	req, metadata, requestHash, model, provider, err := s.prepareCompletion(ctx, cfg, input, structs.RunModeStream)
	if err != nil {
		return nil, err
	}
	req.Stream = true

	client, err := s.manager.get(ctx, cfg, model, nil)
	if err != nil {
		return nil, err
	}

	run, err := s.createRun(ctx, structs.RunModeStream, input.Action, provider, model, requestHash, metadata)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := withPolicyTimeout(ctx, cfg)
	chunks, err := client.Stream(callCtx, req)
	if err != nil {
		cancel()
		_, _ = s.failRun(ctx, cfg, run.ID, time.Now(), err)
		return nil, err
	}

	return &StreamSession{
		Run:         run,
		Chunks:      chunks,
		Config:      cfg,
		Cancel:      cancel,
		StartedAt:   time.Now(),
		Model:       model,
		Provider:    provider,
		RequestHash: requestHash,
	}, nil
}

func (s *aiService) FinishStream(ctx context.Context, session *StreamSession, final *structs.StreamChunkResponse, content strings.Builder, streamErr error) (*structs.Run, error) {
	if session == nil || session.Run == nil {
		return nil, fmt.Errorf("stream session is required")
	}
	if session.Cancel != nil {
		defer session.Cancel()
	}
	if streamErr != nil {
		return s.failRun(ctx, session.Config, session.Run.ID, session.StartedAt, streamErr)
	}
	if final == nil {
		return s.failRun(ctx, session.Config, session.Run.ID, session.StartedAt, io.ErrUnexpectedEOF)
	}

	provider := session.Provider
	model := session.Model
	if provider == "" {
		provider, model, _ = parseProviderModel(session.Model)
	}

	return s.succeedRun(
		ctx,
		session.Config,
		session.Run.ID,
		session.StartedAt,
		provider,
		model,
		session.Config.Model.Primary,
		responseHash(content.String()),
		final.InputTokens,
		final.OutputTokens,
		final.TotalTokens,
		final.ReasoningTokens,
		final.CacheUsage.CreatedTokens,
		final.CacheUsage.ReadTokens,
	)
}

func (s *aiService) Embed(ctx context.Context, input *structs.EmbedRequest) (*structs.EmbedResponse, error) {
	cfg, err := s.manager.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !cfg.Embedding.Enabled {
		return nil, ErrPolicyDenied
	}
	if input == nil || len(input.Input) == 0 {
		return nil, fmt.Errorf("input is required")
	}
	if len(input.Input) > cfg.Embedding.MaxItems {
		return nil, fmt.Errorf("%w: input item count exceeds %d", ErrPolicyDenied, cfg.Embedding.MaxItems)
	}
	totalChars := 0
	for _, value := range input.Input {
		totalChars += len(value)
	}
	if totalChars > cfg.Policy.MaxPromptChars {
		return nil, fmt.Errorf("%w: input length exceeds %d characters", ErrPolicyDenied, cfg.Policy.MaxPromptChars)
	}

	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = cfg.Embedding.Model
	}
	provider, _, err := parseProviderModel(model)
	if err != nil {
		return nil, err
	}

	client, err := s.manager.get(ctx, cfg, model, []string{})
	if err != nil {
		return nil, err
	}

	metadata := map[string]any{
		"input_count": len(input.Input),
		"input_chars": totalChars,
		"input_type":  input.InputType,
	}
	for key, value := range input.Metadata {
		metadata[key] = value
	}
	requestHash := ""
	if cfg.Safety.StoreRequestHash {
		requestHash = hashAny(map[string]any{"input": input.Input, "model": model, "input_type": input.InputType})
	}
	run, err := s.createRun(ctx, structs.RunModeEmbed, "", provider, model, requestHash, metadata)
	if err != nil {
		return nil, err
	}

	callCtx, cancel := withPolicyTimeout(ctx, cfg)
	defer cancel()

	inputType := input.InputType
	if inputType == "" {
		inputType = cfg.Embedding.InputType
	}
	started := time.Now()
	resp, err := client.Embed(callCtx, &deebus.EmbedRequest{
		Input:     input.Input,
		Model:     model,
		InputType: inputType,
	})
	if err != nil {
		_, _ = s.failRun(ctx, cfg, run.ID, started, err)
		return nil, err
	}

	cost, currency := estimateCost(cfg, model, resp.TokensUsed, 0, 0, 0, 0)
	_, err = s.runRepo.Update(ctx, &structs.UpdateRunInput{
		ID:            run.ID,
		Status:        structs.RunStatusSucceeded,
		Provider:      provider,
		Model:         resp.Model,
		InputTokens:   resp.TokensUsed,
		TotalTokens:   resp.TokensUsed,
		DurationMS:    time.Since(started).Milliseconds(),
		EstimatedCost: cost,
		Currency:      currency,
		ResponseHash:  hashAny(map[string]any{"embeddings": len(resp.Embeddings), "model": resp.Model}),
		UpdatedBy:     ctxutil.GetUserID(ctx),
	})
	if err != nil {
		return nil, err
	}

	return &structs.EmbedResponse{
		RunID:      run.ID,
		Provider:   provider,
		Model:      resp.Model,
		Embeddings: resp.Embeddings,
		TokensUsed: resp.TokensUsed,
	}, nil
}

func (s *aiService) RunAction(ctx context.Context, action string, input *structs.ActionRequest) (*structs.ActionResponse, error) {
	if input == nil {
		input = &structs.ActionRequest{}
	}
	action = strings.TrimSpace(action)
	descriptor, ok := actionRegistry[action]
	if !ok {
		return nil, ErrActionNotFound
	}

	cfg, err := s.manager.config.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !isActionAllowed(cfg, action) {
		return nil, ErrPolicyDenied
	}

	complete := buildActionCompletionRequest(descriptor, input)
	complete.Action = action
	complete.Model = input.Model
	complete.MaxOutputTokens = input.MaxOutputTokens
	complete.Temperature = input.Temperature

	resp, err := s.complete(ctx, complete, structs.RunModeAction)
	if err != nil {
		return nil, err
	}

	parsed := parseJSONMap(resp.Content)
	return &structs.ActionResponse{
		RunID:       resp.RunID,
		Action:      action,
		Content:     resp.Content,
		JSON:        parsed,
		Provider:    resp.Provider,
		Model:       resp.Model,
		TotalTokens: resp.TotalTokens,
	}, nil
}

func (s *aiService) prepareCompletion(ctx context.Context, cfg *structs.Config, input *structs.CompleteRequest, mode structs.RunMode) (*deebus.Request, map[string]any, string, string, string, error) {
	if input == nil {
		return nil, nil, "", "", "", fmt.Errorf("request body is required")
	}
	if err := s.validateCompletionPolicy(cfg, input); err != nil {
		return nil, nil, "", "", "", err
	}

	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = cfg.Model.Primary
	}
	provider, _, err := parseProviderModel(model)
	if err != nil {
		return nil, nil, "", "", "", err
	}

	messages := make([]deebus.Message, 0, len(input.Messages)+1)
	if strings.TrimSpace(input.Prompt) != "" {
		messages = append(messages, deebus.TextMessage("user", input.Prompt))
	}
	for _, message := range input.Messages {
		role := strings.TrimSpace(message.Role)
		if role == "" {
			role = "user"
		}
		messages = append(messages, deebus.TextMessage(role, message.Content))
	}
	if len(messages) == 0 {
		return nil, nil, "", "", "", fmt.Errorf("prompt or messages is required")
	}

	maxOutputTokens := input.MaxOutputTokens
	if maxOutputTokens <= 0 {
		maxOutputTokens = cfg.Model.DefaultMaxOutputTokens
	}
	if maxOutputTokens > cfg.Policy.MaxOutputTokens {
		return nil, nil, "", "", "", fmt.Errorf("%w: max_output_tokens exceeds %d", ErrPolicyDenied, cfg.Policy.MaxOutputTokens)
	}

	temperature := cfg.Model.DefaultTemperature
	if input.Temperature != nil {
		temperature = *input.Temperature
	}

	request := &deebus.Request{
		Messages:        messages,
		MaxOutputTokens: maxOutputTokens,
		Temperature:     temperature,
		Stop:            input.Stop,
		Metadata:        input.Metadata,
		UserID:          ctxutil.GetUserID(ctx),
	}
	if input.TopP != nil {
		request.TopP = *input.TopP
	}
	if input.ResponseFormat != nil {
		request.ResponseFormat = &deebus.ResponseFormat{
			Type:        input.ResponseFormat.Type,
			Name:        input.ResponseFormat.Name,
			Description: input.ResponseFormat.Description,
			Schema:      input.ResponseFormat.Schema,
			Strict:      input.ResponseFormat.Strict,
		}
	}
	if input.Reasoning != nil {
		request.Reasoning = &deebus.ReasoningConfig{
			Effort:          input.Reasoning.Effort,
			BudgetTokens:    input.Reasoning.BudgetTokens,
			IncludeThoughts: input.Reasoning.IncludeThoughts,
		}
	}

	metadata := make(map[string]any)
	for key, value := range input.Metadata {
		metadata[key] = value
	}
	metadata["mode"] = string(mode)
	metadata["message_count"] = len(messages)
	metadata["prompt_chars"] = completionCharCount(input)
	metadata["requested_model"] = model
	metadata = sanitizeMetadata(metadata)

	requestHash := ""
	if cfg.Safety.StoreRequestHash {
		requestHash = hashAny(map[string]any{
			"prompt":   input.Prompt,
			"messages": input.Messages,
			"model":    model,
			"action":   input.Action,
		})
	}

	return request, metadata, requestHash, model, provider, nil
}

func (s *aiService) validateCompletionPolicy(cfg *structs.Config, input *structs.CompleteRequest) error {
	if cfg == nil || !cfg.Enabled || !cfg.Policy.Enabled {
		return ErrAIDisabled
	}
	if input.Action != "" && !isActionAllowed(cfg, input.Action) {
		return fmt.Errorf("%w: action %s is not allowed", ErrPolicyDenied, input.Action)
	}
	if len(input.Messages) > cfg.Policy.MaxMessages {
		return fmt.Errorf("%w: message count exceeds %d", ErrPolicyDenied, cfg.Policy.MaxMessages)
	}
	if completionCharCount(input) > cfg.Policy.MaxPromptChars {
		return fmt.Errorf("%w: prompt length exceeds %d characters", ErrPolicyDenied, cfg.Policy.MaxPromptChars)
	}
	for _, message := range input.Messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role == "system" && !cfg.Safety.AllowSystemPrompts {
			return fmt.Errorf("%w: system prompts are disabled", ErrPolicyDenied)
		}
	}
	if len(cfg.Safety.BlockedPhrases) > 0 {
		payload := strings.ToLower(input.Prompt)
		for _, message := range input.Messages {
			payload += "\n" + strings.ToLower(message.Content)
		}
		for _, phrase := range cfg.Safety.BlockedPhrases {
			if phrase != "" && strings.Contains(payload, strings.ToLower(phrase)) {
				return fmt.Errorf("%w: request contains blocked phrase", ErrPolicyDenied)
			}
		}
	}
	return nil
}

func (s *aiService) createRun(ctx context.Context, mode structs.RunMode, action, provider, model, requestHash string, metadata map[string]any) (*structs.Run, error) {
	operationID := ""
	if value, ok := metadata["operation_id"].(string); ok {
		operationID = value
	}
	if operationID == "" {
		operationID = nanoid.String(16)
		metadata["operation_id"] = operationID
	}
	userID := ctxutil.GetUserID(ctx)
	return s.runRepo.Create(ctx, &structs.CreateRunInput{
		OperationID: operationID,
		Action:      action,
		Mode:        mode,
		Status:      structs.RunStatusRunning,
		Provider:    provider,
		Model:       model,
		RequestHash: requestHash,
		Metadata:    metadata,
		SpaceID:     ctxutil.GetSpaceID(ctx),
		UserID:      userID,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	})
}

func (s *aiService) failRun(ctx context.Context, cfg *structs.Config, id string, started time.Time, err error) (*structs.Run, error) {
	code, message := sanitizeError(err, cfg.Safety.MaxErrorChars)
	return s.runRepo.Update(ctx, &structs.UpdateRunInput{
		ID:           id,
		Status:       structs.RunStatusFailed,
		DurationMS:   time.Since(started).Milliseconds(),
		ErrorCode:    code,
		ErrorMessage: message,
		UpdatedBy:    ctxutil.GetUserID(ctx),
	})
}

func (s *aiService) succeedRun(ctx context.Context, cfg *structs.Config, id string, started time.Time, provider, model, configuredPrimary, responseHash string, inputTokens, outputTokens, totalTokens, reasoningTokens, cacheCreated, cacheRead int) (*structs.Run, error) {
	fallback := ""
	if provider != "" && model != "" {
		actualModel := provider + "/" + model
		if configuredPrimary != "" && actualModel != configuredPrimary {
			fallback = actualModel
		}
	}
	cost, currency := estimateCost(cfg, provider+"/"+model, inputTokens, outputTokens, cacheCreated, cacheRead, reasoningTokens)
	return s.runRepo.Update(ctx, &structs.UpdateRunInput{
		ID:                 id,
		Status:             structs.RunStatusSucceeded,
		Provider:           provider,
		Model:              model,
		FallbackModel:      fallback,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		TotalTokens:        totalTokens,
		ReasoningTokens:    reasoningTokens,
		CacheCreatedTokens: cacheCreated,
		CacheReadTokens:    cacheRead,
		DurationMS:         time.Since(started).Milliseconds(),
		ResponseHash:       responseHash,
		EstimatedCost:      cost,
		Currency:           currency,
		UpdatedBy:          ctxutil.GetUserID(ctx),
	})
}

func completionResponse(runID, operationID string, resp *deebus.Response, cfg *structs.Config) *structs.CompleteResponse {
	result := &structs.CompleteResponse{
		RunID:           runID,
		OperationID:     operationID,
		Content:         resp.Content,
		Reasoning:       resp.Reasoning,
		Provider:        resp.Provider,
		Model:           resp.Model,
		FinishReason:    resp.FinishReason,
		InputTokens:     resp.InputTokens,
		OutputTokens:    resp.OutputTokens,
		TotalTokens:     resp.TokensUsed,
		ReasoningTokens: resp.ReasoningTokens,
		CacheUsage: structs.CacheUsage{
			CreatedTokens: resp.CacheUsage.CreatedTokens,
			ReadTokens:    resp.CacheUsage.ReadTokens,
		},
		ToolCalls: deebusToolCalls(resp.ToolCalls),
	}
	if cfg != nil && cfg.Policy.StoreRawOutput {
		result.Raw = resp.Raw
	}
	return result
}

func deebusToolCalls(calls []deebus.ToolCall) []structs.ToolCall {
	if len(calls) == 0 {
		return nil
	}
	results := make([]structs.ToolCall, 0, len(calls))
	for _, call := range calls {
		result := structs.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			Name: call.Function.Name,
		}
		if call.Function.Arguments != "" {
			result.Arguments = json.RawMessage(call.Function.Arguments)
		}
		results = append(results, result)
	}
	return results
}

func buildActionCompletionRequest(descriptor structs.ActionDescriptor, input *structs.ActionRequest) *structs.CompleteRequest {
	contextJSON, _ := json.MarshalIndent(sanitizeMetadata(input.Context), "", "  ")
	language := strings.TrimSpace(input.Language)
	if language == "" {
		language = "same language as the source content unless a locale is explicitly requested"
	}
	tone := strings.TrimSpace(input.Tone)
	if tone == "" {
		tone = "formal, precise, and production-ready"
	}

	system := fmt.Sprintf(
		"You are Ncobase AI for the %s domain. Return only useful production work. Do not claim that you changed system state, do not perform privileged mutations, and do not include secrets. Desired output type: %s.",
		descriptor.Domain,
		descriptor.OutputType,
	)
	user := fmt.Sprintf(
		"Action: %s\nGoal: %s\nLanguage: %s\nTone: %s\nInstruction: %s\nContext:\n%s\nContent:\n%s",
		descriptor.Key,
		descriptor.Description,
		language,
		tone,
		input.Instruction,
		string(contextJSON),
		input.Content,
	)
	req := &structs.CompleteRequest{
		Messages: []structs.MessageInput{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	if descriptor.OutputType == "json" || input.OutputFormat == "json" {
		req.ResponseFormat = &structs.ResponseFormatInput{Type: "json_object"}
	}
	return req
}

func isActionAllowed(cfg *structs.Config, action string) bool {
	allowed := stringSet(cfg.Policy.AllowedActions)
	_, ok := allowed[action]
	return ok
}

func completionCharCount(input *structs.CompleteRequest) int {
	if input == nil {
		return 0
	}
	total := len(input.Prompt)
	for _, message := range input.Messages {
		total += len(message.Content)
	}
	return total
}

func responseHash(content string) string {
	if content == "" {
		return ""
	}
	return hashString(content)
}

func parseJSONMap(content string) map[string]any {
	var result map[string]any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil
	}
	return result
}
