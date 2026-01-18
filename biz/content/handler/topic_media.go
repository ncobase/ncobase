package handler

import (
	"ncobase/biz/content/service"
	"ncobase/biz/content/structs"

	"github.com/ncobase/ncore/validation"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// TopicMediaHandlerInterface is the interface for the handler.
type TopicMediaHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
	GetByTopicAndMedia(c *gin.Context)
	ListByTopic(c *gin.Context)
}

// topicMediaHandler represents the handler.
type topicMediaHandler struct {
	s *service.Service
}

// NewTopicMediaHandler creates a new handler.
func NewTopicMediaHandler(s *service.Service) TopicMediaHandlerInterface {
	return &topicMediaHandler{
		s: s,
	}
}

// Create handles the creation of a topic media relation.
//
// @Summary Create topic media relation
// @Description Create a new relation between a topic and a media resource.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateTopicMediaBody true "CreateTopicMediaBody object"
// @Success 200 {object} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media [post]
// @Security Bearer
func (h *topicMediaHandler) Create(c *gin.Context) {
	body := &structs.CreateTopicMediaBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Validate required fields
	if body.TopicID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("topic_id")))
		return
	}

	if body.MediaID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("media_id")))
		return
	}

	if body.Type == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("type")))
		return
	}

	// Create the topic-media relation
	result, err := h.s.TopicMedia.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a topic media relation.
//
// @Summary Update topic media relation
// @Description Update an existing relation between a topic and media.
// @Tags cms
// @Accept json
// @Produce json
// @Param id path string true "Topic Media ID"
// @Param body body object{TopicID string `json:"topic_id"` MediaID string `json:"media_id"` Type string `json:"type"` Order int `json:"order"`} true "Update parameters"
// @Success 200 {object} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media/{id} [put]
// @Security Bearer
func (h *topicMediaHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var body struct {
		TopicID string `json:"topic_id"`
		MediaID string `json:"media_id"`
		Type    string `json:"type"`
		Order   int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.TopicMedia.Update(c.Request.Context(), id, body.TopicID, body.MediaID, body.Type, body.Order)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a topic media relation.
//
// @Summary Get topic media relation
// @Description Retrieve details of a topic media relation.
// @Tags cms
// @Produce json
// @Param id path string true "Topic Media ID"
// @Success 200 {object} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media/{id} [get]
func (h *topicMediaHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.TopicMedia.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a topic media relation.
//
// @Summary Delete topic media relation
// @Description Delete an existing topic media relation.
// @Tags cms
// @Produce json
// @Param id path string true "Topic Media ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media/{id} [delete]
// @Security Bearer
func (h *topicMediaHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.TopicMedia.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing topic media relations.
//
// @Summary List topic media relations
// @Description Retrieve a list of topic media relations.
// @Tags cms
// @Produce json
// @Param params query structs.ListTopicMediaParams true "List topic media parameters"
// @Success 200 {array} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media [get]
func (h *topicMediaHandler) List(c *gin.Context) {
	params := &structs.ListTopicMediaParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	topicMedia, err := h.s.TopicMedia.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topicMedia)
}

// GetByTopicAndMedia handles getting a topic media relation by topic ID and media ID.
//
// @Summary Get by topic and media
// @Description Retrieve a topic media relation by topic ID and media ID.
// @Tags cms
// @Produce json
// @Param topicId query string true "Topic ID"
// @Param mediaId query string true "Media ID"
// @Success 200 {object} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media/by-topic-and-media [get]
func (h *topicMediaHandler) GetByTopicAndMedia(c *gin.Context) {
	topicID := c.Query("topicId")
	if topicID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("topicId")))
		return
	}

	mediaID := c.Query("mediaId")
	if mediaID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("mediaId")))
		return
	}

	result, err := h.s.TopicMedia.GetByTopicAndMedia(c.Request.Context(), topicID, mediaID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListByTopic handles listing topic media relations for a specific topic.
//
// @Summary List by topic
// @Description Retrieve all media relations for a specific topic.
// @Tags cms
// @Produce json
// @Param topicId path string true "Topic ID"
// @Param type query string false "Media Type" Enums(featured, gallery, attachment)
// @Success 200 {array} structs.ReadTopicMedia "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/topic-media/by-topic/{topicId} [get]
func (h *topicMediaHandler) ListByTopic(c *gin.Context) {
	topicID := c.Param("topicId")
	if topicID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("topicId")))
		return
	}

	params := &structs.ListTopicMediaParams{
		TopicID: topicID,
		Type:    c.Query("type"),
		Limit:   100, // Default large limit to get all media for a topic
	}

	topicMedia, err := h.s.TopicMedia.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, topicMedia)
}
