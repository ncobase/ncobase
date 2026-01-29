package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ncobase/plugin/proxy/data"
	"ncobase/plugin/proxy/data/repository"
	"ncobase/plugin/proxy/structs"
	"strings"
	"text/template"

	"github.com/dop251/goja"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// TransformerFunc is a function type for transformers
type TransformerFunc func([]byte) ([]byte, error)

// TransformerServiceInterface is the interface for the transformer service.
type TransformerServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTransformerBody) (*structs.ReadTransformer, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTransformer, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*structs.ReadTransformer, error)
	GetByName(ctx context.Context, name string) (*structs.ReadTransformer, error)
	List(ctx context.Context, params *structs.ListTransformerParams) (paging.Result[*structs.ReadTransformer], error)
	CompileTransformer(ctx context.Context, id string) (TransformerFunc, error)
}

// transformerService is the struct for the transformer service.
type transformerService struct {
	transformer repository.TransformerRepositoryInterface
}

// NewTransformerService creates a new transformer service.
func NewTransformerService(d *data.Data) TransformerServiceInterface {
	return &transformerService{
		transformer: repository.NewTransformerRepository(d),
	}
}

// Create creates a new transformer.
func (s *transformerService) Create(ctx context.Context, body *structs.CreateTransformerBody) (*structs.ReadTransformer, error) {
	if body.Name == "" {
		return nil, errors.New("transformer name is required")
	}

	// Validate the transformer by attempting to compile it
	_, err := s.compileTransformerFromBody(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("invalid transformer: %w", err)
	}

	row, err := s.transformer.Create(ctx, body)
	if err := handleEntError(ctx, "Transformer", err); err != nil {
		return nil, err
	}

	return repository.SerializeTransformer(row), nil
}

// Update updates an existing transformer.
func (s *transformerService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTransformer, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// If content is being updated, validate the new transformer
	if content, ok := updates["content"].(string); ok {
		// Get the current transformer to check its type
		current, err := s.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}

		// Create a mock body for validation
		mockBody := &structs.CreateTransformerBody{
			TransformerBody: structs.TransformerBody{
				Type:        current.Type,
				Content:     content,
				ContentType: current.ContentType,
			},
		}

		// Override type if it's being updated
		if transformerType, ok := updates["type"].(string); ok {
			mockBody.Type = transformerType
		}

		// Override content_type if it's being updated
		if contentType, ok := updates["content_type"].(string); ok {
			mockBody.ContentType = contentType
		}

		// Validate by compiling
		_, err = s.compileTransformerFromBody(ctx, mockBody)
		if err != nil {
			return nil, fmt.Errorf("invalid transformer: %w", err)
		}
	}

	row, err := s.transformer.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Transformer", err); err != nil {
		return nil, err
	}

	return repository.SerializeTransformer(row), nil
}

// Delete deletes a transformer by ID.
func (s *transformerService) Delete(ctx context.Context, id string) error {
	err := s.transformer.Delete(ctx, id)
	if err := handleEntError(ctx, "Transformer", err); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a transformer by ID.
func (s *transformerService) GetByID(ctx context.Context, id string) (*structs.ReadTransformer, error) {
	row, err := s.transformer.GetByID(ctx, id)
	if err := handleEntError(ctx, "Transformer", err); err != nil {
		return nil, err
	}

	return repository.SerializeTransformer(row), nil
}

// GetByName retrieves a transformer by name.
func (s *transformerService) GetByName(ctx context.Context, name string) (*structs.ReadTransformer, error) {
	row, err := s.transformer.GetByName(ctx, name)
	if err := handleEntError(ctx, "Transformer", err); err != nil {
		return nil, err
	}

	return repository.SerializeTransformer(row), nil
}

// List lists all transformers.
func (s *transformerService) List(ctx context.Context, params *structs.ListTransformerParams) (paging.Result[*structs.ReadTransformer], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTransformer, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.transformer.List(ctx, &lp)
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing transformers: %v", err)
			return nil, 0, err
		}

		total := s.transformer.CountX(ctx, params)

		return repository.SerializeTransformers(rows), total, nil
	})
}

// CompileTransformer compiles a transformer by ID and returns a function that can be used to transform data.
func (s *transformerService) CompileTransformer(ctx context.Context, id string) (TransformerFunc, error) {
	transformer, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.compileTransformer(ctx, transformer)
}

// compileTransformerFromBody compiles a transformer from a create/update body.
func (s *transformerService) compileTransformerFromBody(ctx context.Context, body *structs.CreateTransformerBody) (TransformerFunc, error) {
	// Create a mock transformer object for compilation testing
	transformer := &structs.ReadTransformer{
		Type:        body.Type,
		Content:     body.Content,
		ContentType: body.ContentType,
	}

	return s.compileTransformer(ctx, transformer)
}

// compileTransformer compiles a transformer and returns a function that can be used to transform data.
func (s *transformerService) compileTransformer(ctx context.Context, transformer *structs.ReadTransformer) (TransformerFunc, error) {
	switch transformer.Type {
	case "template":
		return s.compileTemplateTransformer(transformer.Content)
	case "script":
		return s.compileScriptTransformer(transformer.Content)
	case "mapping":
		return s.compileMappingTransformer(transformer.Content)
	default:
		return nil, fmt.Errorf("unsupported transformer type: %s", transformer.Type)
	}
}

// compileTemplateTransformer compiles a template transformer.
func (s *transformerService) compileTemplateTransformer(content string) (TransformerFunc, error) {
	tmpl, err := template.New("transformer").Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return func(data []byte) ([]byte, error) {
		// Parse the input data
		var input any
		if err := json.Unmarshal(data, &input); err != nil {
			// If not valid JSON, use it as a string
			input = string(data)
		}

		// Apply the template
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, input); err != nil {
			return nil, fmt.Errorf("template execution failed: %w", err)
		}

		return buf.Bytes(), nil
	}, nil
}

// compileScriptTransformer compiles a JavaScript transformer.
func (s *transformerService) compileScriptTransformer(content string) (TransformerFunc, error) {
	// Pre-check the script
	vm := goja.New()
	_, err := vm.RunString(content)
	if err != nil {
		return nil, fmt.Errorf("invalid JavaScript: %w", err)
	}

	// The transform function should be defined in the script
	return func(data []byte) ([]byte, error) {
		vm := goja.New()

		// Define the input
		err := vm.Set("input", string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to set input: %w", err)
		}

		// Try to parse input as JSON if possible
		if _, err := vm.RunString(`
			try {
				input = JSON.parse(input);
			} catch (e) {
				// If not JSON, keep as string
			}
		`); err != nil {
			return nil, fmt.Errorf("failed to prepare input: %w", err)
		}

		// Run the transformer script
		_, err = vm.RunString(content)
		if err != nil {
			return nil, fmt.Errorf("script execution failed: %w", err)
		}

		// Get the result
		result, err := vm.RunString("JSON.stringify(transform(input))")
		if err != nil {
			return nil, fmt.Errorf("transform function not found or failed: %w", err)
		}

		return []byte(result.String()), nil
	}, nil
}

// compileMappingTransformer compiles a mapping transformer.
func (s *transformerService) compileMappingTransformer(content string) (TransformerFunc, error) {
	// Parse the mapping configuration
	var mapping struct {
		Mappings []struct {
			Source       string `json:"source"`
			Target       string `json:"target"`
			DefaultValue any    `json:"default_value,omitempty"`
			Transform    string `json:"transform,omitempty"`
		} `json:"mappings"`
	}

	if err := json.Unmarshal([]byte(content), &mapping); err != nil {
		return nil, fmt.Errorf("invalid mapping configuration: %w", err)
	}

	return func(data []byte) ([]byte, error) {
		// Parse the input data
		var input map[string]any
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("input is not valid JSON: %w", err)
		}

		// Apply the mapping
		result := make(map[string]any)
		for _, m := range mapping.Mappings {
			var value any

			// Get the source value
			if strings.Contains(m.Source, ".") {
				// Handle nested paths
				parts := strings.Split(m.Source, ".")
				nested := input
				found := true

				for _, part := range parts[:len(parts)-1] {
					if nestedObj, ok := nested[part].(map[string]any); ok {
						nested = nestedObj
					} else {
						found = false
						break
					}
				}

				if found {
					value = nested[parts[len(parts)-1]]
				}
			} else {
				value = input[m.Source]
			}

			// Use default value if source is not found
			if value == nil && m.DefaultValue != nil {
				value = m.DefaultValue
			}

			// Apply transformation if specified
			if m.Transform != "" && value != nil {
				vm := goja.New()
				err := vm.Set("value", value)
				if err != nil {
					return nil, fmt.Errorf("failed to set value for transformation: %w", err)
				}

				result, err := vm.RunString(m.Transform)
				if err != nil {
					return nil, fmt.Errorf("transformation failed: %w", err)
				}

				value = result.Export()
			}

			// Set the target value
			if strings.Contains(m.Target, ".") {
				// Handle nested paths
				parts := strings.Split(m.Target, ".")
				nested := result

				// Create nested structure
				for _, part := range parts[:len(parts)-1] {
					if _, ok := nested[part]; !ok {
						nested[part] = make(map[string]any)
					}
					nested = nested[part].(map[string]any)
				}

				nested[parts[len(parts)-1]] = value
			} else {
				result[m.Target] = value
			}
		}

		return json.Marshal(result)
	}, nil
}
