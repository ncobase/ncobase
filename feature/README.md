# Features for Ncobase

> This is the plug-in and module in the Next-Gen system.

## Structure

```plantext
├── access/               # Access control and permission management
├── auth/                 # Authentication and authorization
├── content/              # Content management
├── group/                # Group management
├── init/                 # System initialization
├── relationship/         # Relationship management (consider merging into other modules)
├── resource/             # Resource management
├── system/               # System-related feature, built-in
├── tenant/               # Tenant management
├── user/                 # User management
├── asset/                # Asset plugin directory
├── interface.go          # Feature interface
├── manager.go            # Feature manager
├── plugin.go             # Plugin manager
└── README.md             # README
```

Each directory contains its own README.md file explaining its structure and functionality.
