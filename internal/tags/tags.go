package tags

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewPGXRepository),
	do.Lazy(NewService),
)
