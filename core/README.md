# Core Domain Modules


> Core domains provide fundamental infrastructure and services for the entire system.


## Structure

```plantext
├── access/               # Access control and permission management, including role, permission, casbin, role permission
├── auth/                 # Authentication and authorization
├── realtime/             # Real-time management, Used for notification, message, etc.
├── space/                # organization structure, and team management, including group, department, member, role
├── system/               # Core system functionalities
├── tenant/               # Multi-tenancy support
├── user/                 # User management and profiles
├── workflow/             # Workflow management, Used for business process, such as approval, workflow design, etc.
└── README.md             # This file
```

Each directory contains its own README.md file explaining its structure and functionality.
