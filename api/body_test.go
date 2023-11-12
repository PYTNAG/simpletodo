package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type requestBody gin.H

func (body requestBody) replace(key string, newValue any) requestBody {
	newBody := make(requestBody, len(body))

	if _, ok := body[key]; !ok {
		panic(fmt.Errorf("body doesn't have field %s", key))
	}

	for field, value := range body {
		if field == key {
			newBody[field] = newValue
			continue
		}

		newBody[field] = value
	}

	return newBody
}
