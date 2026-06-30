package service

import (
	"context"
	"errors"
	"fmt"
	contentStructs "ncobase/biz/content/structs"
	resourceEnt "ncobase/plugin/resource/data/ent"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/structs"
	"testing"

	"github.com/ncobase/ncore/paging"
	"github.com/ncobase/ncore/types"
)

type fakeFileRepository struct {
	files     map[string]*resourceEnt.File
	deletedID string
}

func (r *fakeFileRepository) Create(context.Context, *structs.CreateFileBody) (*resourceEnt.File, error) {
	return nil, errors.New("not implemented")
}

func (r *fakeFileRepository) GetByID(_ context.Context, slug string) (*resourceEnt.File, error) {
	if file, ok := r.files[slug]; ok {
		return file, nil
	}
	return nil, fmt.Errorf("file %s not found", slug)
}

func (r *fakeFileRepository) GetByHash(context.Context, string, string) (*resourceEnt.File, error) {
	return nil, errors.New("not implemented")
}

func (r *fakeFileRepository) Update(context.Context, string, types.JSON) (*resourceEnt.File, error) {
	return nil, errors.New("not implemented")
}

func (r *fakeFileRepository) Delete(_ context.Context, slug string) error {
	if _, ok := r.files[slug]; !ok {
		return fmt.Errorf("file %s not found", slug)
	}
	r.deletedID = slug
	delete(r.files, slug)
	return nil
}

func (r *fakeFileRepository) List(context.Context, *structs.ListFileParams) ([]*resourceEnt.File, error) {
	return nil, errors.New("not implemented")
}

func (r *fakeFileRepository) CountX(context.Context, *structs.ListFileParams) int { return 0 }

func (r *fakeFileRepository) SumSizeByOwner(context.Context, string) (int64, error) { return 0, nil }

func (r *fakeFileRepository) SumSizeBySpace(context.Context, string) (int64, error) { return 0, nil }

func (r *fakeFileRepository) GetAllOwners(context.Context) ([]string, error) { return nil, nil }

func (r *fakeFileRepository) GetAllSpaces(context.Context) ([]string, error) { return nil, nil }

func (r *fakeFileRepository) AggregateSizeByCategory(context.Context) (map[string]int64, error) {
	return nil, nil
}

func (r *fakeFileRepository) AggregateSizeByStorage(context.Context) (map[string]int64, error) {
	return nil, nil
}

func (r *fakeFileRepository) AggregateDailyUsage(context.Context, int64, int64) ([]repository.DailyUsageAggregate, error) {
	return nil, nil
}

func (r *fakeFileRepository) AggregateOwnerUsage(context.Context, int) ([]repository.OwnerUsageAggregate, error) {
	return nil, nil
}

func (r *fakeFileRepository) AggregateUsageBetween(context.Context, int64, int64) (int64, int, error) {
	return 0, 0, nil
}

func (r *fakeFileRepository) SearchByTags(context.Context, string, []string, int) ([]*resourceEnt.File, error) {
	return nil, nil
}

func (r *fakeFileRepository) GetTagsByOwner(context.Context, string) ([]string, error) {
	return nil, nil
}

func (r *fakeFileRepository) CheckNameExists(context.Context, string, string) (bool, error) {
	return false, nil
}

func (r *fakeFileRepository) FindExpiredFiles(context.Context, *structs.CleanupFilters, int) ([]*resourceEnt.File, error) {
	return nil, nil
}

func (r *fakeFileRepository) FindOrphanedFiles(context.Context, *structs.CleanupFilters, int) ([]*resourceEnt.File, error) {
	return nil, nil
}

func (r *fakeFileRepository) FindDuplicateFiles(context.Context) (map[string][]*resourceEnt.File, error) {
	return nil, nil
}

type fakeContentResolver struct {
	available        bool
	mediaByFile      map[string][]*contentStructs.ReadMedia
	relationsByMedia map[string][]*contentStructs.ReadTopicMedia
	topics           map[string]*contentStructs.ReadTopic
	mediaErr         error
}

func (r *fakeContentResolver) HasContentServices() bool {
	return r.available
}

func (r *fakeContentResolver) ListMedia(_ context.Context, params *contentStructs.ListMediaParams) (paging.Result[*contentStructs.ReadMedia], error) {
	if r.mediaErr != nil {
		return paging.Result[*contentStructs.ReadMedia]{}, r.mediaErr
	}
	items := append([]*contentStructs.ReadMedia{}, r.mediaByFile[params.ResourceID]...)
	return paging.Result[*contentStructs.ReadMedia]{
		Items: items,
		Total: len(items),
	}, nil
}

func (r *fakeContentResolver) ListTopicMedia(_ context.Context, params *contentStructs.ListTopicMediaParams) (paging.Result[*contentStructs.ReadTopicMedia], error) {
	items := append([]*contentStructs.ReadTopicMedia{}, r.relationsByMedia[params.MediaID]...)
	return paging.Result[*contentStructs.ReadTopicMedia]{
		Items: items,
		Total: len(items),
	}, nil
}

func (r *fakeContentResolver) GetTopic(_ context.Context, id string) (*contentStructs.ReadTopic, error) {
	if topic, ok := r.topics[id]; ok {
		return topic, nil
	}
	return nil, fmt.Errorf("topic %s not found", id)
}

func testFileService(repo *fakeFileRepository, content contentReferenceResolver) *fileService {
	return &fileService{
		fileRepo: repo,
		content:  content,
	}
}

func resourceFile(id string) *resourceEnt.File {
	return &resourceEnt.File{
		ID:           id,
		Name:         id,
		OriginalName: id + ".png",
		Path:         "uploads/" + id + ".png",
		Type:         "image/png",
		Size:         1024,
		OwnerID:      "user-1",
		AccessLevel:  string(structs.AccessLevelPrivate),
		Extras:       map[string]any{"space_id": "space-1"},
	}
}

func TestDeleteImpactAllowsUnreferencedResourceDeletion(t *testing.T) {
	repo := &fakeFileRepository{files: map[string]*resourceEnt.File{"file-1": resourceFile("file-1")}}
	content := &fakeContentResolver{available: true, mediaByFile: map[string][]*contentStructs.ReadMedia{}}
	svc := testFileService(repo, content)

	impact, err := svc.DeleteImpact(context.Background(), []string{"file-1"})
	if err != nil {
		t.Fatalf("DeleteImpact returned error: %v", err)
	}
	if impact.Summary == nil || !impact.Summary.CanDelete {
		t.Fatalf("expected delete to be allowed, got %#v", impact.Summary)
	}

	if err := svc.Delete(context.Background(), "file-1"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if repo.deletedID != "file-1" {
		t.Fatalf("expected file-1 to be deleted, got %q", repo.deletedID)
	}
}

func TestDeleteImpactBlocksReferencedResourceDeletion(t *testing.T) {
	repo := &fakeFileRepository{files: map[string]*resourceEnt.File{"file-1": resourceFile("file-1")}}
	content := &fakeContentResolver{
		available: true,
		mediaByFile: map[string][]*contentStructs.ReadMedia{
			"file-1": {
				{ID: "media-1", Title: "Hero", Type: "image", ResourceID: "file-1", SpaceID: "space-1"},
			},
		},
		relationsByMedia: map[string][]*contentStructs.ReadTopicMedia{
			"media-1": {
				{ID: "topic-media-1", MediaID: "media-1", TopicID: "topic-1", Type: "featured", Order: 1},
			},
		},
		topics: map[string]*contentStructs.ReadTopic{
			"topic-1": {ID: "topic-1", Title: "Launch", Slug: "launch", SpaceID: "space-1"},
		},
	}
	svc := testFileService(repo, content)

	impact, err := svc.DeleteImpact(context.Background(), []string{"file-1"})
	if err != nil {
		t.Fatalf("DeleteImpact returned error: %v", err)
	}
	if impact.Summary == nil || impact.Summary.CanDelete {
		t.Fatalf("expected delete to be blocked, got %#v", impact.Summary)
	}
	if got := impact.Summary.MediaReferenceCount; got != 1 {
		t.Fatalf("expected one media reference, got %d", got)
	}
	if got := impact.Summary.TopicReferenceCount; got != 1 {
		t.Fatalf("expected one topic reference, got %d", got)
	}
	if impact.Impacts[0].TopicReferences[0].Topic.Title != "Launch" {
		t.Fatalf("expected topic summary to be attached, got %#v", impact.Impacts[0].TopicReferences[0].Topic)
	}

	err = svc.Delete(context.Background(), "file-1")
	var blocked *ResourceDeleteBlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("expected ResourceDeleteBlockedError, got %v", err)
	}
	if repo.deletedID != "" {
		t.Fatalf("referenced file should not be deleted, got %q", repo.deletedID)
	}
}

func TestDeleteImpactBlocksWhenReferenceCheckFails(t *testing.T) {
	repo := &fakeFileRepository{files: map[string]*resourceEnt.File{"file-1": resourceFile("file-1")}}
	content := &fakeContentResolver{available: true, mediaErr: errors.New("media query failed")}
	svc := testFileService(repo, content)

	impact, err := svc.DeleteImpact(context.Background(), []string{"file-1"})
	if err != nil {
		t.Fatalf("DeleteImpact returned error: %v", err)
	}
	if impact.Summary == nil || impact.Summary.CanDelete {
		t.Fatalf("expected failed check to block deletion, got %#v", impact.Summary)
	}
	if impact.Summary.ErrorCount != 1 {
		t.Fatalf("expected one error, got %d", impact.Summary.ErrorCount)
	}

	err = svc.Delete(context.Background(), "file-1")
	var blocked *ResourceDeleteBlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("expected ResourceDeleteBlockedError, got %v", err)
	}
}

func TestDeleteImpactRejectsTooManyFiles(t *testing.T) {
	repo := &fakeFileRepository{files: map[string]*resourceEnt.File{}}
	content := &fakeContentResolver{available: true}
	svc := testFileService(repo, content)

	ids := make([]string, 0, structs.DeleteImpactMaxFiles+1)
	for i := 0; i <= structs.DeleteImpactMaxFiles; i++ {
		ids = append(ids, fmt.Sprintf("file-%d", i))
	}

	if _, err := svc.DeleteImpact(context.Background(), ids); err == nil {
		t.Fatalf("expected max file count error")
	}
}

func TestDeleteMissingFileIsNotReportedAsReferenceBlock(t *testing.T) {
	repo := &fakeFileRepository{files: map[string]*resourceEnt.File{}}
	content := &fakeContentResolver{available: true}
	svc := testFileService(repo, content)

	err := svc.Delete(context.Background(), "missing-file")
	if err == nil {
		t.Fatalf("expected missing file error")
	}
	var blocked *ResourceDeleteBlockedError
	if errors.As(err, &blocked) {
		t.Fatalf("missing file should not be reported as a reference block")
	}
}
