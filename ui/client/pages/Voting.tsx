import { useState, useEffect, useRef } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ArrowLeft, Heart, X, RotateCcw, Clock, Star } from "lucide-react";
import { toast } from "sonner";

interface MovieMeta {
  id: string;
  title: string;
  poster: string;
  duration: string;
  genre: string[];
  description: string;
  year: number;
  rating: number;
}

export default function Voting() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const roomId = searchParams.get("room") || "";

  const [MovieMetas, setMovieMetas] = useState<MovieMeta[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isFlipped, setIsFlipped] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [votes, setVotes] = useState<Record<string, "like" | "dislike">>({});

  const cardRef = useRef<HTMLDivElement>(null);
  const startPos = useRef({ x: 0, y: 0 });

  // Mock MovieMeta data - will be replaced with API call
  useEffect(() => {
    // TODO: API call to get MovieMetas for voting
    // const response = await fetch(`/api/rooms/${roomId}/MovieMetas`);
    // const data = await response.json();
    // setMovieMetas(data.MovieMetas);

    // Mock data for demonstration
    const mockMovieMetas: MovieMeta[] = [
      {
        id: "MovieMeta-1",
        title: "Маска",
        poster: "/placeholder.svg",
        duration: "101 мин",
        genre: ["Комедия", "Фэнтези"],
        description:
          "Застенчивый банковский служащий Стэнли Ипкисс находит древнюю маску, которая превращает его в зеленолицого супергероя с невероятными способностями.",
        year: 1994,
        rating: 6.9,
      },
      {
        id: "MovieMeta-2",
        title: "Джон Уик",
        poster: "/placeholder.svg",
        duration: "101 мин",
        genre: ["Боевик", "Триллер"],
        description:
          "Легендарный киллер выходит из отставки, чтобы отомстить за убийство своей собаки - последнего подарка от покойной жены.",
        year: 2014,
        rating: 7.4,
      },
      {
        id: "MovieMeta-3",
        title: "Интерстеллар",
        poster: "/placeholder.svg",
        duration: "169 мин",
        genre: ["Фантастика", "Драма"],
        description:
          "В будущем Земля умирает, и группа исследователей отправляется через червоточину в поисках нового дома для человечества.",
        year: 2014,
        rating: 8.6,
      },
    ];

    setMovieMetas(mockMovieMetas);
  }, [roomId]);

  const currentMovieMeta = MovieMetas[currentIndex];

  const handleVote = async (vote: "like" | "dislike") => {
    if (!currentMovieMeta) return;

    // TODO: API call to submit vote
    // await fetch(`/api/rooms/${roomId}/vote`, {
    //   method: 'POST',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify({ MovieMetaId: currentMovieMeta.id, vote })
    // });

    setVotes((prev) => ({ ...prev, [currentMovieMeta.id]: vote }));

    if (vote === "like") {
      toast.success("Фильм понравился!");
    } else {
      toast.success("Фильм не понравился");
    }

    // Move to next MovieMeta
    if (currentIndex < MovieMetas.length - 1) {
      setCurrentIndex((prev) => prev + 1);
      setIsFlipped(false);
    } else {
      // All MovieMetas voted - go to results
      navigate(`/results?room=${roomId}`);
    }
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    setIsDragging(true);
    startPos.current = { x: e.clientX, y: e.clientY };
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!isDragging) return;

    const deltaX = e.clientX - startPos.current.x;
    const deltaY = e.clientY - startPos.current.y;

    setDragOffset({ x: deltaX, y: deltaY });
  };

  const handleMouseUp = () => {
    if (!isDragging) return;

    const threshold = 100;

    if (Math.abs(dragOffset.x) > threshold) {
      const vote = dragOffset.x > 0 ? "like" : "dislike";
      handleVote(vote);
    }

    setIsDragging(false);
    setDragOffset({ x: 0, y: 0 });
  };

  // Touch events for mobile
  const handleTouchStart = (e: React.TouchEvent) => {
    const touch = e.touches[0];
    setIsDragging(true);
    startPos.current = { x: touch.clientX, y: touch.clientY };
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!isDragging) return;

    const touch = e.touches[0];
    const deltaX = touch.clientX - startPos.current.x;
    const deltaY = touch.clientY - startPos.current.y;

    setDragOffset({ x: deltaX, y: deltaY });
  };

  const handleTouchEnd = () => {
    handleMouseUp();
  };

  if (!currentMovieMeta) {
    return (
      <div className="min-h-screen bg-background p-4 flex items-center justify-center">
        <Card className="p-8 max-w-md w-full text-center space-y-6 bg-card border-border">
          <div className="w-16 h-16 mx-auto bg-acid-green/10 rounded-full flex items-center justify-center">
            <Heart className="w-8 h-8 text-acid-green" />
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl font-bold text-foreground">
              Голосование завершено!
            </h1>
            <p className="text-muted-foreground">
              Ожидайте результатов от других участников
            </p>
          </div>
        </Card>
      </div>
    );
  }

  const cardStyle = {
    transform: `translate(${dragOffset.x}px, ${dragOffset.y}px) rotate(${dragOffset.x * 0.1}deg)`,
    opacity: Math.max(0.7, 1 - Math.abs(dragOffset.x) / 300),
  };

  return (
    <div className="min-h-screen bg-background p-4 overflow-hidden">
      <div className="max-w-md mx-auto h-full flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate(-1)}
            className="text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Назад
          </Button>

          <div className="text-center">
            <h1 className="text-xl font-bold text-foreground">Голосование</h1>
            <p className="text-sm text-muted-foreground">
              {currentIndex + 1} из {MovieMetas.length}
            </p>
          </div>

          <div className="w-16" />
        </div>

        {/* MovieMeta Card */}
        <div className="flex-1 flex items-center justify-center mb-6">
          <div className="relative w-full max-w-sm">
            <div
              ref={cardRef}
              className="MovieMeta-card cursor-grab active:cursor-grabbing"
              style={cardStyle}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseUp}
              onTouchStart={handleTouchStart}
              onTouchMove={handleTouchMove}
              onTouchEnd={handleTouchEnd}
              onClick={() => setIsFlipped(!isFlipped)}
            >
              <Card className="relative h-[600px] w-full bg-card border-border overflow-hidden transform-gpu transition-transform duration-300 preserve-3d">
                <div
                  className={`absolute inset-0 backface-hidden ${isFlipped ? "rotate-y-180" : ""}`}
                >
                  {/* Front of card */}
                  <div className="h-full flex flex-col">
                    {/* MovieMeta Poster */}
                    <div className="relative flex-1 bg-muted">
                      <img
                        src={currentMovieMeta.poster}
                        alt={currentMovieMeta.title}
                        className="w-full h-full object-cover"
                      />
                      <div className="absolute top-4 right-4 bg-black/70 text-white px-2 py-1 rounded-lg text-sm font-medium">
                        <div className="flex items-center space-x-1">
                          <Star className="w-3 h-3 text-yellow-400 fill-current" />
                          <span>{currentMovieMeta.rating}</span>
                        </div>
                      </div>
                    </div>

                    {/* MovieMeta Info */}
                    <div className="p-6 space-y-3">
                      <div>
                        <h2 className="text-2xl font-bold text-foreground">
                          {currentMovieMeta.title}
                        </h2>
                        <p className="text-muted-foreground">
                          {currentMovieMeta.year}
                        </p>
                      </div>

                      <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                        <div className="flex items-center space-x-1">
                          <Clock className="w-4 h-4" />
                          <span>{currentMovieMeta.duration}</span>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        {currentMovieMeta.genre.map((g) => (
                          <span
                            key={g}
                            className="px-3 py-1 bg-accent text-accent-foreground rounded-full text-sm"
                          >
                            {g}
                          </span>
                        ))}
                      </div>

                      <p className="text-center text-sm text-muted-foreground">
                        Нажмите, чтобы посмотреть описание
                      </p>
                    </div>
                  </div>
                </div>

                <div
                  className={`absolute inset-0 backface-hidden rotate-y-180 ${isFlipped ? "rotate-y-0" : ""}`}
                >
                  {/* Back of card */}
                  <div className="h-full flex flex-col p-6 bg-card">
                    <div className="flex items-center justify-between mb-6">
                      <h3 className="text-xl font-bold text-foreground">
                        {currentMovieMeta.title}
                      </h3>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          setIsFlipped(false);
                        }}
                      >
                        <RotateCcw className="w-4 h-4" />
                      </Button>
                    </div>

                    <div className="flex-1 overflow-auto">
                      <p className="text-foreground leading-relaxed">
                        {currentMovieMeta.description}
                      </p>
                    </div>

                    <div className="mt-6 pt-4 border-t border-border">
                      <div className="text-sm text-muted-foreground space-y-1">
                        <p>
                          <strong>Год:</strong> {currentMovieMeta.year}
                        </p>
                        <p>
                          <strong>Жанр:</strong>{" "}
                          {currentMovieMeta.genre.join(", ")}
                        </p>
                        <p>
                          <strong>Длительность:</strong>{" "}
                          {currentMovieMeta.duration}
                        </p>
                        <p>
                          <strong>Рейтинг:</strong> {currentMovieMeta.rating}/10
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            </div>

            {/* Swipe indicators */}
            {isDragging && (
              <>
                <div
                  className={`absolute top-1/2 left-4 transform -translate-y-1/2 text-6xl transition-opacity ${
                    dragOffset.x > 50
                      ? "opacity-100 text-acid-green"
                      : "opacity-30 text-muted-foreground"
                  }`}
                >
                  ❤️
                </div>
                <div
                  className={`absolute top-1/2 right-4 transform -translate-y-1/2 text-6xl transition-opacity ${
                    dragOffset.x < -50
                      ? "opacity-100 text-red-500"
                      : "opacity-30 text-muted-foreground"
                  }`}
                >
                  ❌
                </div>
              </>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex justify-center space-x-8 mb-6">
          <Button
            variant="outline"
            size="lg"
            onClick={() => handleVote("dislike")}
            className="w-16 h-16 rounded-full border-red-500 text-red-500 hover:bg-red-500 hover:text-white transition-colors"
          >
            <X className="w-6 h-6" />
          </Button>

          <Button
            variant="outline"
            size="lg"
            onClick={() => handleVote("like")}
            className="w-16 h-16 rounded-full border-acid-green text-acid-green hover:bg-acid-green hover:text-acid-green-foreground transition-colors"
          >
            <Heart className="w-6 h-6" />
          </Button>
        </div>

        {/* Progress */}
        <div className="w-full bg-muted rounded-full h-2">
          <div
            className="bg-acid-green h-2 rounded-full transition-all duration-300"
            style={{
              width: `${((currentIndex + 1) / MovieMetas.length) * 100}%`,
            }}
          />
        </div>
      </div>
    </div>
  );
}
