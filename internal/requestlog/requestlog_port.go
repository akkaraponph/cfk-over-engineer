package requestlog

import "github.com/samber/mo"

type Repository interface {
	FindByID(id string) mo.Option[RequestLog]
	FindAll(limit, offset int) mo.Result[[]RequestLog]
}
