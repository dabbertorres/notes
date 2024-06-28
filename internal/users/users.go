package users

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewService),
	do.Lazy(NewPGXRepository),
)
