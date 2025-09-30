import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import { Film, Users, Copy } from 'lucide-react';
import { toast } from 'sonner';

export default function Index() {
  const [roomCode, setRoomCode] = useState(['', '', '', '', '', '']);
  const [isJoinDisabled, setIsJoinDisabled] = useState(true);
  const navigate = useNavigate();
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleDigitChange = (index: number, value: string) => {
    // Only allow single digits
    if (value.length > 1) return;
    if (value && !/^\d$/.test(value)) return;

    const newCode = [...roomCode];
    newCode[index] = value;
    setRoomCode(newCode);

    // Check if all digits are filled
    setIsJoinDisabled(newCode.some(digit => !digit));

    // Auto-focus next input
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    // Handle backspace
    if (e.key === 'Backspace' && !roomCode[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pasteData = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const newCode = ['', '', '', '', '', ''];

    for (let i = 0; i < pasteData.length && i < 6; i++) {
      newCode[i] = pasteData[i];
    }

    setRoomCode(newCode);
    setIsJoinDisabled(newCode.some(digit => !digit));

    // Focus the next empty input or the last one
    const nextEmptyIndex = newCode.findIndex(digit => !digit);
    const targetIndex = nextEmptyIndex !== -1 ? nextEmptyIndex : Math.min(pasteData.length, 5);
    inputRefs.current[targetIndex]?.focus();
  };

  const handleCreateRoom = async () => {
    try {
      const response = await fetch('/api/rooms', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        }
      });

      if (!response.ok) {
        throw new Error('Failed to create room');
      }

      const data = await response.json();
      navigate(`/rooms/${data.room_id}/lobby?host=true`);
    } catch (error) {
      console.error('Error creating room:', error);
      toast.error('Не удалось создать комнату. Попробуйте еще раз.');
    }
  };

  const handleJoinRoom = async () => {
    if (isJoinDisabled) return;

    const roomId = Array.isArray(roomCode) ? roomCode.join('') : roomCode;
    try {
      const response = await fetch(`/api/rooms/${roomId}/acquired`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        }
      });

      if (response.status === 404) {
        toast.error('Комната недоступна, попробуйте другой код');
        return;
      }

      if (!response.ok) {
        throw new Error('Failed to check room');
      }

      // Room exists, navigate to lobby
      navigate(`/rooms/${roomId}/lobby`);
    } catch (error) {
      console.error('Error joining room:', error);
      toast.error('Ошибка при подключении к комнате. Попробуйте еще раз.');
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8">
        {/* Logo and Header */}
        <div className="text-center space-y-4">
          <div className="flex items-center justify-center mb-6">
            <div className="relative">
              <Film className="w-16 h-16 text-acid-green" />
              <div className="absolute -top-2 -right-2 w-6 h-6 bg-acid-green rounded-full flex items-center justify-center">
                <Users className="w-3 h-3 text-black" />
              </div>
            </div>
          </div>
          
          <h1 className="text-4xl font-bold text-foreground">
            Kino<span className="text-acid-green">Swap</span>
          </h1>
          
          <p className="text-muted-foreground text-lg">
            Выбирайте фильмы вместе с друзьями
          </p>
        </div>

        {/* Main Actions */}
        <div className="space-y-6">
          {/* Create Room Button */}
          <Button 
            onClick={handleCreateRoom}
            className="w-full h-14 text-lg font-semibold bg-acid-green hover:bg-acid-green/90 text-acid-green-foreground rounded-xl transition-all duration-200 transform hover:scale-[1.02]"
          >
            <Film className="w-5 h-5 mr-2" />
            Создать голосование
          </Button>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-border"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="bg-background px-4 text-muted-foreground">или</span>
            </div>
          </div>

          {/* Join Room Section */}
          <Card className="p-6 space-y-4 bg-card border-border">
            <h3 className="text-lg font-semibold text-foreground text-center">
              Присоединиться к голосованию
            </h3>
            
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm text-muted-foreground text-center block">
                  Код комнаты
                </label>
                <div className="flex justify-center space-x-2" onPaste={handlePaste}>
                  {roomCode.map((digit, index) => (
                    <div key={index} className="relative">
                      <Input
                        ref={(el) => (inputRefs.current[index] = el)}
                        value={digit}
                        onChange={(e) => handleDigitChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        className="w-12 h-12 text-center text-xl font-mono bg-input border-border focus:ring-acid-green focus:border-acid-green rounded-xl"
                        maxLength={1}
                        inputMode="numeric"
                        pattern="[0-9]"
                      />
                      {index === 2 && (
                        <div className="absolute -right-4 top-1/2 transform -translate-y-1/2 text-muted-foreground text-xl font-mono">
                          -
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
              
              <Button 
                onClick={handleJoinRoom}
                disabled={isJoinDisabled}
                className="w-full h-12 bg-primary hover:bg-primary/90 text-primary-foreground disabled:opacity-50 disabled:cursor-not-allowed rounded-xl transition-all duration-200"
              >
                <Users className="w-4 h-4 mr-2" />
                Присоединиться
              </Button>
            </div>
          </Card>
        </div>

        {/* Footer */}
        <div className="text-center pt-8">
          <p className="text-sm text-muted-foreground">
            Для использования в Telegram Mini App
          </p>
        </div>
      </div>
    </div>
  );
}
