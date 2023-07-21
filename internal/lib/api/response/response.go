package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOk  = "OK"
	StatusErr = "ERR"
)

func Ok() Response {
	return Response{
		Status: StatusOk,
	}
}

func Err(err string) Response {
	return Response{
		Status: StatusErr,
		Error:  err,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errorList []string

	for _, err := range errs {
        switch err.ActualTag() {
			case "required":
				errorList = append(errorList, fmt.Sprintf("field %s is a required field", err.Field()))
			default:
				errorList = append(errorList, fmt.Sprintf("field %s is invalid", err.Field()))
		}
    }

	return Response{
		Status: StatusErr,
		Error:  strings.Join(errorList, ", "),
	}
}