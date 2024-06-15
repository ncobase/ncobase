//go:build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

// TODO: implement
// // wireApp is a Wire provider function that initializes the server.
// func wireApp(conf *config.Config) (*server.Server, func(), error) {
// 	// Initialize the logger
// 	loggerClean, err := log.Init(conf.Logger)
// 	if err != nil {
// 		return nil, nil, err
// 	}
//
// 	// Create and return the server instance along with the cleanup function
// 	serve, err := server.New(conf)
// 	if err != nil {
// 		loggerClean() // Clean up logger before returning
// 		return nil, nil, err
// 	}
// 	return serve, loggerClean, nil
// }
