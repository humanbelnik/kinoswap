# Backend Integration Notes

## Successfully Integrated Endpoints

✅ **GET /api/rooms** - Create room (acquireRoom)
- Used in: Index.tsx `handleCreateRoom`
- Returns: `{"room_id": "string"}`

✅ **GET /api/rooms/:room_id/acquired** - Check if room exists (isRoomAcquired)  
- Used in: Index.tsx `handleJoinRoom`
- Returns: 200 OK or 404 Not Found

✅ **PATCH /api/rooms/:room_id/participate** - Submit preferences (participate)
- Used in: Lobby.tsx `handleReady`
- Body: `{"text": "user preferences"}`
- Returns: 200 OK

✅ **PATCH /api/rooms/:room_id/start** - Start voting (start)
- Used in: Lobby.tsx `handleStartVoting` (host only)
- Returns: 200 OK

## Additional Endpoint Needed for Complete Functionality

To implement the auto-redirect feature when voting starts, you'll need to add one more endpoint:

### GET /api/rooms/:room_id/status
**Purpose:** Check room status and whether voting has started  
**Response:**
```json
{
  "room_id": "123456",
  "voting_started": false,
  "participants_count": 3,
  "ready_participants": 2
}
```

**Usage:** 
- Called every 2 seconds by non-host users in the lobby
- When `voting_started` becomes `true`, all participants auto-redirect to voting page

## Current Polling Implementation

The frontend currently polls the `/acquired` endpoint every 2 seconds for non-host users. Once you add the `/status` endpoint, uncomment the TODO section in `client/pages/Lobby.tsx` around line 41-46.

## Error Handling

- 404 responses show user-friendly error messages
- 400 responses indicate validation errors  
- 500 responses show generic error messages
- All errors are displayed using toast notifications

## Notes

- Host users don't poll for status since they control voting start
- Room codes are automatically formatted as XXX-XXX in the UI
- All API calls include proper error handling and user feedback
