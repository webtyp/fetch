# Architecture Analysis: fetchgo vs CRUDP Alignment

## Executive Summary

**Conclusión:** `fetchgo` y `CRUDP` son **complementarios pero no idénticos**. `fetchgo` debe ser una librería HTTP genérica de bajo nivel, y `CRUDP` debe usar `fetchgo` como su capa de transporte.

---

## 1. Respuesta a: "¿SetBaseURL en el Cliente o Manager?"

### Problema Identificado

La arquitectura actual propone:
```go
client := fetchgo.New().NewClient("https://api1.com", 5000)
client.SendRequest("POST", "/users", body, callback)
```

**Limitaciones:**
- ❌ Un cliente = un dominio
- ❌ Para múltiples APIs necesitas múltiples clientes
- ❌ No puedes hacer peticiones a terceros (ej. OAuth, webhooks)

### Solución Propuesta: URLs Absolutas

```go
// ✅ Arquitectura Recomendada
api := fetchgo.New()
client := api.NewClient(5000) // Sin baseURL

// Ahora puedes usar cualquier dominio
client.SendRequest("POST", "https://api1.com/users", "application/json", data, cb)
client.SendRequest("POST", "https://api2.com/orders", "application/json", data, cb)
client.SendRequest("GET", "https://oauth.google.com/token", "application/json", nil, cb)
```

**Beneficios:**
- ✅ Un solo cliente para todo (como `fetch` del browser)
- ✅ Compatible con microservicios
- ✅ Compatible con CRUDP (que puede enviar a múltiples endpoints)

**Si necesitas un "baseURL" por conveniencia:**

Crea un wrapper ligero en tu app:

```go
// En tu aplicación (no en fetchgo)
type APIClient struct {
    baseURL string
    client  fetchgo.Client
}

func (a *APIClient) Post(endpoint string, data any, cb func(any, error)) {
    fullURL := a.baseURL + endpoint
    a.client.SendRequest("POST", fullURL, "application/json", data, cb)
}
```

---

## 2. Alineación con CRUDP: Tabla Comparativa

| Característica | CRUDP (Spec) | fetchgo (Actual) | ¿Compatible? | Acción Requerida |
|----------------|--------------|------------------|--------------|------------------|
| **Múltiples Dominios** | ✅ Necesario (sync endpoint + SSE endpoint) | ❌ baseURL fijo | **No** | Eliminar baseURL del cliente |
| **Protocolo Binario** | ✅ TinyBin packets | ⚠️ HTTP con body binario | **Parcial** | fetchgo ya soporta TinyBin |
| **Batching** | ✅ `[]Packet` en una request | ❌ 1 request = 1 operación | **No** | Crear layer CRUDP sobre fetchgo |
| **Async (SSE)** | ✅ Respuesta por eventos | ❌ Callback inmediato | **No** | CRUDP debe gestionar SSE aparte |
| **Codec Auto** | ✅ Por Content-Type | ✅ Por Content-Type | **Sí** | ✅ Ya compatible |
| **CORS** | ✅ En manager | ✅ En manager | **Sí** | ✅ Ya compatible |
| **Offline Queue** | ✅ `Enqueue()` → `Sync()` | ❌ Envío inmediato | **No** | Implementar en CRUDP client |

---

## 3. Respuesta: "¿Debe fetchgo Implementar CRUDP?"

### ❌ No. Mantenerlos Separados

**Razones arquitectónicas:**

1. **Single Responsibility Principle (SRP):**
   - `fetchgo`: HTTP client multiplataforma (como `axios` o `fetch`)
   - `crudp`: Protocolo de sincronización con batching y SSE (como `Apollo Client` o `tRPC`)

2. **Reutilización:**
   ```
   fetchgo (genérico)
      ↑
      ├── crudp-client (tu spec)
      ├── otro-proyecto-rest
      └── graphql-client
   ```

3. **Testing:**
   - Puedes testear `fetchgo` independientemente (con httptest)
   - Puedes testear `crudp` mockeando `fetchgo`

### ✅ Arquitectura de Capas Propuesta

```
┌─────────────────────────────────────────┐
│  UI / Business Logic                    │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│  CRUDP Client (crudp-client.go)         │
│  - Enqueue(handlerID, action, data)     │
│  - Sync() → BatchRequest                │
│  - ListenSSE() → BatchResponse          │
│  - Manejo de ReqID/callbacks            │
└─────────────┬───────────────────────────┘
              │ usa
┌─────────────▼───────────────────────────┐
│  fetchgo (HTTP layer)                   │
│  - SendRequest(method, url, contentType)│
│  - Codecs automáticos (JSON/TinyBin)    │
│  - CORS, Timeouts                       │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│  Platform (stdlib / WASM)               │
│  - net/http  OR  syscall/js (fetch)     │
└─────────────────────────────────────────┘
```

---

## 4. API Simplificada para `fetchgo` (Final)

### Interface Minimalista

```go
// types.go
package fetchgo

// Client define la interfaz pública del cliente HTTP.
type Client interface {
    // SendRequest realiza una petición HTTP completa.
    // url DEBE ser absoluta (ej: "https://api.com/users")
    // contentType determina el codec (ej: "application/json", "application/x-tinybin")
    SendRequest(method, url, contentType string, body any, callback func(any, error))
    
    // SetHeader configura headers por defecto (ej: Authorization)
    // Es chainable para permitir: client.SetHeader("X", "Y").SetHeader("Z", "W")
    SetHeader(key, value string) Client
}
```

### Uso Real

```go
// Crear el manager (una vez en tu app)
http := fetchgo.New().SetCORS("cors", true)

// Crear cliente (timeout 5s)
client := http.NewClient(5000)

// Configurar headers comunes
client.SetHeader("Authorization", "Bearer token123").
       SetHeader("X-App-Version", "1.0")

// JSON a tu backend (auto-detectado porque userData es un struct)
client.SendRequest("POST", "https://api.tuapp.com/users", 
    userData, func(resp any, err error) {
        if err != nil { /* error */ }
        // resp ya viene decodificado automáticamente
    })

// TinyBin a otro servicio (auto-detectado porque binaryData es []byte)
client.SendRequest("POST", "https://internal.tuapp.com/process", 
    binaryData, callback)

// GET sin body
client.SendRequest("GET", "https://api.google.com/data", 
    nil, callback)
```

---

## 5. Cómo CRUDP Usaría fetchgo

```go
// pkg/crudp/client.go
package crudp

import "webtyp.com/fetch"

type Client struct {
    http      fetchgo.Client
    syncURL   string // https://api.tuapp.com/sync
    sseURL    string // https://api.tuapp.com/events
    queue     []Packet
    pending   map[string]ResponseCallback
}

func NewCRUDPClient(baseURL string) *Client {
    httpClient := fetchgo.New().NewClient(5000)
    
    return &Client{
        http:    httpClient,
        syncURL: baseURL + "/sync",
        sseURL:  baseURL + "/events",
        queue:   make([]Packet, 0),
        pending: make(map[string]ResponseCallback),
    }
}

// Enqueue NO usa la red todavía
func (c *Client) Enqueue(handlerID uint8, action byte, callback ResponseCallback, data ...any) {
    reqID := generateID()
    
    packet := Packet{
        ReqID: reqID,
        HandlerID: handlerID,
        Action: action,
        Data: encodeAll(data),
    }
    
    c.queue = append(c.queue, packet)
    c.pending[reqID] = callback
}

// Sync envía todo usando fetchgo
func (c *Client) Sync() {
    batch := BatchRequest{Packets: c.queue}
    c.queue = nil
    
    // ✅ Usa fetchgo para el transporte
    c.http.SendRequest("POST", c.syncURL, "application/x-tinybin", 
        batch, func(resp any, err error) {
            if err != nil {
                // Reintentar, guardar en IndexedDB, etc.
            }
        })
}

// ListenSSE usa la API nativa del browser (no fetchgo)
func (c *Client) ListenSSE() {
    // syscall/js con EventSource directamente
    // Porque SSE no es un request/response estándar
}
```

---

## 6. Cambios Requeridos en la Documentación

### `01_ARCHITECTURE.md`

- [x] ~~Eliminar `baseURL` del `Client`~~
- [x] ~~Cambiar firma: `NewClient(timeoutMS int) Client`~~
- [x] ~~Simplificar interfaz a solo `SendRequest()` y `SetHeader()`~~
- [x] ~~Documentar que URLs deben ser absolutas~~

### `02_IMPLEMENTATION_GUIDE.md`

- [ ] Actualizar checklist de implementación
- [ ] Agregar sección "Integration with CRUDP"

### `03_TESTING_PLAN.md`

- [ ] Agregar test: "Multiple domains in single client"
- [ ] Agregar test: "Absolute URLs validation"

---

## 7. Conclusión y Recomendación Final

### Para `fetchgo`:

✅ **Hacer:**
- Mantenerlo como HTTP client genérico
- Eliminar `baseURL` del cliente
- API minimalista: `SendRequest(method, url, contentType, body, callback)`
- Codec automático por `Content-Type`

❌ **No hacer:**
- No implementar batching (eso es de CRUDP)
- No implementar SSE listening (eso es de CRUDP)
- No implementar cola offline (eso es de CRUDP)

### Para CRUDP:

✅ **Crear un nuevo paquete `crudp-client`** que:
- Use `fetchgo` para el transporte HTTP
- Implemente `Enqueue()`, `Sync()`, `ListenSSE()`
- Gestione el mapa de callbacks por `ReqID`
- Maneje la cola offline

---

**¿Esta arquitectura de dos capas te hace sentido?** Si estás de acuerdo, actualizo todos los documentos para reflejar esta decisión arquitectónica.
