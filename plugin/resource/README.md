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
└── config/             # Runtime option defaults
```

## Configuration

Runtime policy is loaded from system options, not `config.yaml`.

- `resource.upload`: maximum upload size, allowed file types, default storage label.
- `resource.image`: thumbnail defaults, resize limits, compression quality.
- `resource.quota`: quota enablement, enforcement, default quota, warning threshold, check interval.

Object storage provider, endpoint, bucket, and credentials remain infrastructure config in
`config.yaml` through ncore/oss.

## API Endpoints

### Files

- `GET /res` - List files
- `POST /res` - Create file
- `GET /res/:slug` - Get file details
- `GET /res/:slug/delete-impact` - Check deletion impact for one file
- `PUT /res/:slug` - Update file
- `POST /res/delete-impact` - Check deletion impact for up to 100 unique files
- `DELETE /res/:slug` - Delete file after reference checks pass

### Batch Operations

- `POST /res/batch/upload` - Batch upload files
- `POST /res/batch/process` - Batch process files

### Quota Management

- `GET /res/quota` - Get current quota summary
- `GET /res/usage` - Get current usage statistics
