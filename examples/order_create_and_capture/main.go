package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tavocg/go-paypal"
)

func main() {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	if clientID == "" {
		log.Fatal("PAYPAL_CLIENT_ID is required")
	}

	client, err := paypal.NewClient(paypal.SandboxHost)
	if err != nil {
		log.Fatalf("create paypal client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler(clientID))

	mux.HandleFunc("POST /api/orders", createOrderHandler(client))
	mux.HandleFunc("POST /api/orders/{orderID}/capture", captureOrderHandler(client))

	addr := envDefault("ADDR", ":8080")
	log.Printf("serving PayPal sandbox example on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func indexHandler(clientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, indexHTML, clientID)
	}
}

func createOrderHandler(client *paypal.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		response, err := client.CreateOrder(
			ctx,
			"USD",
			"10.00",
			paypal.WithOrderImmediatePayment(),
			paypal.WithoutOrderShipping(),
			paypal.WithOrderPaypalCountry("CR"),
		)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}

		writeJSON(w, http.StatusCreated, response)
	}
}

func captureOrderHandler(client *paypal.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.PathValue("orderID")
		if orderID == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("orderID is required"))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		response, err := client.CaptureOrderPayment(ctx, orderID)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>PayPal Order Create and Capture</title>
  <script src="https://www.paypal.com/sdk/js?client-id=%s&currency=USD"></script>
  <style>
    :root {
      color-scheme: light;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: #17202a;
      background: #f5f7fb;
    }

    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 24px;
    }

    main {
      width: min(100%%, 520px);
      background: #ffffff;
      border: 1px solid #dde4f0;
      border-radius: 8px;
      padding: 28px;
      box-shadow: 0 16px 40px rgb(23 32 42 / 8%%);
    }

    h1 {
      margin: 0 0 8px;
      font-size: 1.5rem;
      line-height: 1.2;
    }

    p {
      margin: 0 0 22px;
      color: #566476;
    }

    #message {
      min-height: 24px;
      margin-top: 18px;
      font-size: 0.95rem;
      color: #25364a;
      word-break: break-word;
    }
  </style>
</head>
<body>
  <main>
    <h1>PayPal Sandbox Checkout</h1>
    <p>Creates and captures a $10.00 USD sandbox order through this Go server.</p>
    <div id="paypal-button-container"></div>
    <div id="message" role="status" aria-live="polite"></div>
  </main>

  <script>
    const message = document.getElementById("message");

    function showMessage(text) {
      message.textContent = text;
    }

    paypal.Buttons({
      async createOrder() {
        const response = await fetch("/api/orders", { method: "POST" });
        const order = await response.json();

        if (!response.ok) {
          throw new Error(order.error || "Failed to create order");
        }

        return order.id;
      },

      async onApprove(data) {
        const response = await fetch("/api/orders/" + data.orderID + "/capture", { method: "POST" });
        const details = await response.json();

        if (!response.ok) {
          throw new Error(details.error || "Failed to capture order");
        }

        const capture = details.purchase_units?.[0]?.payments?.captures?.[0];
        showMessage("Capture " + (capture?.id || details.id) + " completed with status " + (capture?.status || "unknown") + ".");
      },

      onCancel() {
        showMessage("Checkout canceled.");
      },

      onError(err) {
        console.error(err);
        showMessage(err.message || "Checkout failed.");
      }
    }).render("#paypal-button-container");
  </script>
</body>
</html>
`
