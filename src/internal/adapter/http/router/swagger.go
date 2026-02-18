package router

import (
	"fmt"
	"net/http"
)

func registerSwaggerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, swaggerHTML, "/swagger/openapi.json")
	})

	mux.HandleFunc("/swagger/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(openAPI))
	})
}

const swaggerHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>CCY Payment Processor API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "%s",
        dom_id: "#swagger-ui"
      });
    };
  </script>
</body>
</html>`

const openAPI = `{
  "openapi": "3.0.3",
  "info": {
    "title": "CCY Payment Processor API",
    "version": "1.0.0"
  },
  "paths": {
    "/accounts": {
      "post": {
        "summary": "Create account",
        "security": [
          {
            "ChannelID": [],
            "ChannelKey": []
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["customerId", "currency"],
                "properties": {
                  "customerId": {"type": "string"},
                  "currency": {"type": "string", "enum": ["USD", "EUR", "GBP"]},
                  "initialDeposit": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {"description": "Created"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/users": {
      "post": {
        "summary": "Create user",
        "security": [
          {
            "ChannelID": [],
            "ChannelKey": []
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": [
                  "firstName",
                  "lastName",
                  "dob",
                  "phoneNumber",
                  "idType",
                  "idNumber",
                  "kycLevel",
                  "transactionPinHas"
                ],
                "properties": {
                  "firstName": {"type": "string"},
                  "middleName": {"type": "string"},
                  "lastName": {"type": "string"},
                  "dob": {"type": "string", "example": "1992-01-01"},
                  "phoneNumber": {"type": "string"},
                  "idType": {"type": "string", "enum": ["Passport", "DL"]},
                  "idNumber": {"type": "string"},
                  "kycLevel": {"type": "integer", "minimum": 1},
                  "transactionPinHas": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {"description": "Created"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "ChannelID": {
        "type": "apiKey",
        "in": "header",
        "name": "X-Channel-ID"
      },
      "ChannelKey": {
        "type": "apiKey",
        "in": "header",
        "name": "X-Channel-Key"
      }
    }
  }
}`
