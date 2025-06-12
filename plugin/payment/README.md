# Payment

A flexible payment processing module that supports multiple payment providers, subscriptions, and multi-space.

## Features

- **Multiple Payment Providers**: Stripe, PayPal, Alipay, and WeChat Pay support out of the box
- **Subscription Management**: Create and manage recurring subscriptions with trial periods
- **Multi-space Support**: Isolate payment configurations per space
- **Event System**: Rich event system for integrating with other modules
- **RESTful API**: Well-documented API endpoints with Swagger annotations
- **Extensible Architecture**: Easily add new payment providers

## Architecture

The module follows a clean architecture pattern:

- **Handlers**: API endpoints for interacting with the payment system
- **Services**: Business logic layer
- **Repositories**: Data access layer
- **Events**: Pub/sub event system for inter-module communication
- **Providers**: Payment provider implementations

## API Endpoints

### Channels

- `GET /pay/channels` - List payment channels
- `POST /pay/channels` - Create a new payment channel
- `GET /pay/channels/:id` - Get a payment channel
- `PUT /pay/channels/:id` - Update a payment channel
- `DELETE /pay/channels/:id` - Delete a payment channel
- `PUT /pay/channels/:id/status` - Change channel status

### Orders

- `GET /pay/orders` - List payment orders
- `POST /pay/orders` - Create a new payment order
- `GET /pay/orders/:id` - Get a payment order
- `GET /pay/orders/number/:orderNumber` - Get an order by order number
- `POST /pay/orders/:id/payment-url` - Generate payment URL
- `POST /pay/orders/:id/verify` - Verify payment
- `POST /pay/orders/:id/refund` - Refund payment

### Products

- `GET /pay/products` - List products
- `POST /pay/products` - Create a new product
- `GET /pay/products/:id` - Get a product
- `PUT /pay/products/:id` - Update a product
- `DELETE /pay/products/:id` - Delete a product

### Subscriptions

- `GET /pay/subscriptions` - List subscriptions
- `POST /pay/subscriptions` - Create a new subscription
- `GET /pay/subscriptions/:id` - Get a subscription
- `PUT /pay/subscriptions/:id` - Update a subscription
- `POST /pay/subscriptions/:id/cancel` - Cancel a subscription
- `GET /pay/subscriptions/user/:userId` - Get user subscriptions

### Webhooks

- `POST /pay/webhooks/:channel` - Process provider webhooks

### Utility

- `GET /pay/providers` - List available payment providers
- `GET /pay/stats` - Get payment statistics

## Integration

To use this module:

1. Register the module in your application
2. Configure payment providers
3. Create payment channels
4. Create products for sale
5. Create orders and subscriptions

## Example Usage

```go
// Initialize the payment module
paymentModule := payment.New()
paymentModule.Init(conf, extensionManager)

// Register routes
r := gin.New()
paymentModule.RegisterRoutes(r.Group("/api"))
```

## License

MIT
