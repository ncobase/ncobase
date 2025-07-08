# Resource Plugin

A comprehensive file management plugin for the ncore framework.

## Features

- File upload/download with multiple storage backends
- Image processing and thumbnail generation
- Batch operations for multiple files
- Storage quota management
- File versioning and access control
- Full-text search and tag-based filtering

## Structure

```
├── data/               # Data layer with ent ORM
├── handler/            # HTTP request handlers
├── service/            # Business logic services
├── structs/            # Data structures and models
├── event/              # Event publishing and handling
└── config.go           # Configuration management
```

## Configuration

The plugin supports various configuration options for storage, image processing, and quota management. See config.go for details.

## API Endpoints

### Files

- `GET /res` - List files
- `POST /res` - Create file
- `GET /res/:slug` - Get file details
- `PUT /res/:slug` - Update file
- `DELETE /res/:slug` - Delete file

### Batch Operations

- `POST /res/batch/upload` - Batch upload files
- `POST /res/batch/process` - Batch process files

### Quota Management

- `GET /res/quotas` - Get quota
- `PUT /res/quotas` - Set quota
- `GET /res/quotas/usage` - Get usage statistics
