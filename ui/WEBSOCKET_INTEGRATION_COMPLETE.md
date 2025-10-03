# WebSocket Integration Complete Guide

## 🚀 Frontend Implementation (Already Done)

The frontend WebSocket client is implemented in `client/pages/Lobby.tsx` and handles:

### WebSocket Connection
- Connects to: `ws://localhost:8080/api/rooms/{room_id}/ws`
- Automatically handles connection errors and reconnection
- Proper cleanup on component unmount

### Message Handling
```javascript
case 'participant_joined':
  // Updates participant count for host users
  
case 'participant_ready':  
  // Updates ready participants count for host users
  
case 'voting_started':
  // Auto-redirects ALL users to voting page
  // Shows success toast notification
```

## 🔧 Backend Implementation Required

Follow the steps in `WEBSOCKET_BACKEND_GUIDE.md`:

### 1. Install Dependencies
```bash
go get github.com/gorilla/websocket
```

### 2. Add WebSocket Endpoint
```go
rooms.GET("/:room_id/ws", c.handleWebSocket)
```

### 3. Key Integration Points

**When user clicks "Я готов" (participate):**
```go
// After successful participation
c.hub.BroadcastToRoom(roomID, "participant_ready", map[string]interface{}{
    "ready_participants": newReadyCount,
})
```

**When host clicks "Начать голосование" (start):**
```go
// After successful start
c.hub.BroadcastToRoom(roomID, "voting_started", map[string]interface{}{
    "message": "voting started",
})
```

## 🎯 How Auto-Redirect Works

1. **Host clicks "Начать голосование"**
2. **Backend receives PATCH `/api/rooms/{id}/start`**
3. **Backend broadcasts `voting_started` message to all WebSocket clients in room**
4. **ALL participants (including host) receive WebSocket message**
5. **Frontend automatically redirects everyone to `/voting?room={id}`**
6. **Success toast shows "Голосование началось!"**

## 🔄 Message Flow Diagram

```
Host clicks "Start" → Backend API → WebSocket Broadcast → All Clients → Auto Redirect
                                        ↓
                              ┌─────────────────────────┐
                              │ participant_ready       │
                              │ voting_started         │  
                              │ participant_joined     │
                              └─────────────────────────┘
```

## 🛠️ Testing

1. **Open multiple browser tabs/windows**
2. **Create room in one tab (becomes host)**
3. **Join room with code in other tabs**
4. **Click "Я готов" in participant tabs**
5. **Click "Начать голосование" in host tab**
6. **Watch all tabs auto-redirect to voting page**

## 🐛 Error Handling

- **Connection errors**: Toast notification + console logging
- **Message parsing errors**: Graceful handling with console logs
- **Abnormal disconnection**: User notification
- **Failed API calls**: Error toasts with retry suggestion

## 📝 Next Steps

1. Implement the WebSocket backend following `WEBSOCKET_BACKEND_GUIDE.md`
2. Test the real-time auto-redirect functionality
3. Optional: Add participant count updates via WebSocket
4. Optional: Add typing indicators or other real-time features

The frontend is ready and will work immediately once you implement the backend WebSocket endpoint!
