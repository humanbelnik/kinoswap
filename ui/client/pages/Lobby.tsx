import { useState, useEffect, useRef } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { Copy, Users, Play, Check, ArrowLeft } from 'lucide-react';
import { toast } from 'sonner';

export default function Lobby() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { room_id: roomId } = useParams<{ room_id: string }>();
  const isHost = searchParams.get('host') === 'true';
  
  const [participants, setParticipants] = useState(1);
  const [preferences, setPreferences] = useState('');
  const [isReady, setIsReady] = useState(false);
  const [readyParticipants, setReadyParticipants] = useState(0);

  // Format room code as XXX-XXX
  const formattedRoomCode = roomId;

// useEffect(() => {
//   const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
//   const ws = new WebSocket(`${protocol}//${window.location.host}/api/rooms/${roomId}/ws`);

//   ws.onopen = () => console.log('WS connected');
//   ws.onmessage = (event) => {
//     const message = JSON.parse(event.data);

//     if (message.type === 'participant_ready') {
//       setReadyParticipants(message.data.ready_participants);
//     }

//     if (message.type === 'voting_started') {
//       toast.success('Голосование началось!');
//       navigate(`/voting?room=${roomId}`);
//     }
//   };

//   return () => ws.close();
// }, [roomId, navigate]);

  const copyRoomCode = async () => {
    try {
      await navigator.clipboard.writeText(formattedRoomCode);
      toast.success('Код комнаты скопирован!');
    } catch (err) {
      toast.error('Не удалось скопировать код');
    }
  };
const handleReady = async () => {
  console.log('handleReady called');
  console.log('Preferences:', preferences);
  console.log('Room ID:', roomId);
  console.log(JSON.stringify({ preferences }));
  try {
    const response = await fetch(`/api/rooms/${roomId}/participate`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: preferences })
    });

    console.log('Response status:', response.status);
    console.log('Response ok:', response.ok);

    if (!response.ok) {
      const errorText = await response.text();
      console.error('Server error response:', errorText);
      throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
    }

    const result = await response.json();
    console.log('Success response:', result);

    setIsReady(true);
    setReadyParticipants(prev => prev + 1);
    toast.success(preferences ? 'Ваши пожелания отправлены!' : 'Вы подтвердили участие без пожеланий');
    
  } catch (err) {
    console.error('Error in handleReady:', err);
    if (err instanceof Error) {
      toast.error(`Ошибка: ${err.message}`);
    } else {
      toast.error('Не удалось отправить пожелания. Попробуйте еще раз.');
    }
  }
};

  const handleStartVoting = async () => {
    try {
      const response = await fetch(`/api/rooms/${roomId}/start`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' }
      });

      if (!response.ok) {
        throw new Error('Failed to start voting');
      }

      // WebSocket will handle the navigation and toast notification
      // No need to manually navigate here since WebSocket message will trigger it
    } catch (error) {
      console.error('Error starting voting:', error);
      toast.error('Не удалось начать голосование. Попробуйте еще раз.');
    }
  };

  const handleLeaveRoom = () => {
    navigate('/');
  };

  return (
    <div className="min-h-screen bg-background p-4">
      <div className="max-w-md mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleLeaveRoom}
            className="text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Назад
          </Button>
          
          <div className="text-center">
            <h1 className="text-xl font-bold text-foreground">
              Комната готовится
            </h1>
          </div>
          
          <div className="w-16" /> {/* Spacer for alignment */}
        </div>

        {/* Room Code Card */}
        <Card className="p-6 bg-card border-border">
          <div className="text-center space-y-4">
            <h2 className="text-lg font-semibold text-foreground">
              Код комнаты
            </h2>
            
            <div className="flex items-center justify-center space-x-2">
              <div className="text-3xl font-mono font-bold text-acid-green bg-acid-green/10 px-4 py-2 rounded-lg border border-acid-green/20">
                {formattedRoomCode}
              </div>
              
              <Button
                variant="ghost"
                size="sm"
                onClick={copyRoomCode}
                className="text-acid-green hover:bg-acid-green/10"
              >
                <Copy className="w-4 h-4" />
              </Button>
            </div>
            
            <p className="text-sm text-muted-foreground">
              Поделитесь этим кодом с друзьями
            </p>
          </div>
        </Card>

        {/* Participants - Only visible to host */}
        {isHost && (
          <Card className="p-6 bg-card border-border">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <Users className="w-5 h-5 text-acid-green" />
                <span className="text-foreground font-medium">Участники</span>
              </div>
              <div className="text-right">
                <div className="text-2xl font-bold text-acid-green">{participants}</div>
                {readyParticipants > 0 && (
                  <div className="text-xs text-muted-foreground">
                    {readyParticipants} готовы
                  </div>
                )}
              </div>
            </div>
          </Card>
        )}

        {/* Preferences */}
        <Card className="p-6 bg-card border-border space-y-4">
          <h3 className="text-lg font-semibold text-foreground">
            Ваши пожелания
          </h3>
          
          <div className="space-y-2">
            <label htmlFor="preferences" className="text-sm text-muted-foreground">
              Опишите, что хотели бы посмотреть
            </label>
          <Textarea
            id="preferences"
            value={preferences}
            onChange={(e) => setPreferences(e.target.value)}
            placeholder="Например: комедия с Джимом Керри..."
            className="min-h-[120px] bg-input border-border focus:ring-acid-green focus:border-acid-green resize-none"
            disabled={isReady}
/>
          </div>
          
          {!isReady ? (
            <Button
              type="button"
              onClick={handleReady}
              className="w-full h-12 bg-acid-green hover:bg-acid-green/90 text-acid-green-foreground rounded-xl transition-all duration-200"
            >
              <Check className="w-4 h-4 mr-2" />
              Я готов
            </Button>
          ) : (
            <div className="flex items-center justify-center p-3 bg-acid-green/10 rounded-xl border border-acid-green/20">
              <Check className="w-5 h-5 text-acid-green mr-2" />
              <span className="text-acid-green font-medium">Готов к голосованию</span>
            </div>
          )}
        </Card>

        {/* Start Voting Button (Host Only) */}
        {isHost && (
          <Card className="p-6 bg-card border-border">
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-foreground text-center">
                Управление комнатой
              </h3>
              
              <Button
                onClick={handleStartVoting}
                className="w-full h-14 bg-primary hover:bg-primary/90 text-primary-foreground rounded-xl transition-all duration-200 text-lg font-semibold"
              >
                <Play className="w-5 h-5 mr-2" />
                Начать голосование
              </Button>

              <p className="text-sm text-muted-foreground text-center">
                {readyParticipants > 0
                  ? `${readyParticipants} участников готовы`
                  : 'Вы можете начать голосование в любой момент'
                }
              </p>
            </div>
          </Card>
        )}

        {/* Info */}
        <div className="text-center pt-4">
          <p className="text-sm text-muted-foreground">
            {isHost 
              ? 'Вы создатель комнаты. Начните голосование, когда все будут готовы.' 
              : 'Дождитесь, пока создатель комнаты начнет голосов��ние.'
            }
          </p>
        </div>
      </div>
    </div>
  );
}
