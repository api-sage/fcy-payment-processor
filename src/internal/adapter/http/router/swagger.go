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
  <title>FCY Payment Processor API Docs</title>
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
    "title": "FCY Payment Processor API",
    "version": "1.0.0"
  },
  "paths": {
    "/create-account": {
      "post": {
        "summary": "Create account",
        "security": [
          {
            "BasicAuth": []
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
    "/get-account": {
      "get": {
        "summary": "Get account by account number",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "parameters": [
          {
            "name": "accountNumber",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string",
              "pattern": "^[0-9]{10}$"
            }
          },
          {
            "name": "bankCode",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string",
              "pattern": "^[0-9]{6}$"
            }
          }
        ],
        "responses": {
          "200": {"description": "Account fetched"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/get-participant-banks": {
      "get": {
        "summary": "Get participant banks",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "responses": {
          "200": {"description": "Participant banks fetched"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/get-rates": {
      "get": {
        "summary": "Get all rates",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "responses": {
          "200": {"description": "Rates fetched"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/get-rate": {
      "post": {
        "summary": "Get a specific rate by currency pair",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["fromCurrency", "toCurrency"],
                "properties": {
                  "fromCurrency": {"type": "string", "example": "USD"},
                  "toCurrency": {"type": "string", "example": "NGN"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Rate fetched"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/get-charges": {
      "get": {
        "summary": "Get charge and VAT breakdown",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "parameters": [
          {
            "name": "amount",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string",
              "example": "100.00"
            }
          },
          {
            "name": "fromCurrency",
            "in": "query",
            "required": true,
            "schema": {
              "type": "string",
              "example": "USD"
            }
          }
        ],
        "responses": {
          "200": {"description": "Charges fetched"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/convertfcyamount": {
      "post": {
        "summary": "Get converted amount by currency pair",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["amount", "fromCcy", "toCcy"],
                "properties": {
                  "amount": {"type": "string", "example": "1500.00"},
                  "fromCcy": {"type": "string", "example": "USD"},
                  "toCcy": {"type": "string", "example": "NGN"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Converted rate fetched"},
          "400": {"description": "Validation error"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    },
    "/create-user": {
      "post": {
        "summary": "Create user",
        "security": [
          {
            "BasicAuth": []
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
                  "transactionPin"
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
                  "transactionPin": {"type": "string"}
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
    "/verify-pin": {
      "post": {
        "summary": "Verify user transaction pin",
        "security": [
          {
            "BasicAuth": []
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["customerId", "pin"],
                "properties": {
                  "customerId": {"type": "string"},
                  "pin": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Pin verified"},
          "400": {"description": "Validation error or invalid pin"},
          "401": {"description": "Unauthorized"},
          "500": {"description": "Server error"}
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BasicAuth": {
        "type": "http",
        "scheme": "basic"
      }
    }
  }
}`
