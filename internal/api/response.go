package api

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *ListMeta   `json:"meta,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

type ListMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ErrorData struct {
	Number  int    `json:"number"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *fiber.Ctx, data interface{}, status ...int) error {
	s := fiber.StatusOK
	if len(status) > 0 {
		s = status[0]
	}
	return c.Status(s).JSON(Response{
		Success: true,
		Data:    data,
	})
}

func SuccessList(c *fiber.Ctx, data interface{}, meta ListMeta, status ...int) error {
	s := fiber.StatusOK
	if len(status) > 0 {
		s = status[0]
	}
	return c.Status(s).JSON(Response{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *fiber.Ctx, code string, message string, number int, status ...int) error {
	s := fiber.StatusInternalServerError
	if len(status) > 0 {
		s = status[0]
	}
	return c.Status(s).JSON(Response{
		Success: false,
		Error: &ErrorData{
			Number:  number,
			Code:    code,
			Message: message,
		},
	})
}

func SendError(c *fiber.Ctx, err ApiError, status ...int) error {
	return Error(c, err.Code, err.Message, err.Number, status...)
}
