package http

import (
	"apihorpug/docs"

	"github.com/gofiber/fiber/v3"
)

const scalarPageHTML = `<!doctype html>
<html>
  <head>
    <title>Horpug API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/swagger/doc.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

const swaggerPageHTML = `<!doctype html>
<html>
  <head>
    <title>Horpug API Docs</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: "/swagger/doc.json",
          dom_id: "#swagger-ui",
        });
      };
    </script>
  </body>
</html>`

func RegisterDocsRoutes(app *fiber.App) {
	app.Get("/swagger/doc.json", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})

	app.Get("/docs/swagger", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(swaggerPageHTML)
	})

	app.Get("/docs/scalar", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(scalarPageHTML)
	})
}
