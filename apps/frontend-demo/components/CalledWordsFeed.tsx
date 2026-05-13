import { CalledWord } from "@/types/game"

interface CalledWordsFeedProps {
  words: CalledWord[]
}

export function CalledWordsFeed({ words }: CalledWordsFeedProps) {
  const sortedWords = [...words].sort((a, b) => new Date(b.calledAt).getTime() - new Date(a.calledAt).getTime());

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <h3 className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-2 sm:mb-4 shrink-0">Recent Calls</h3>
      <div className="flex flex-wrap gap-2 overflow-y-auto pb-4">
        {sortedWords.map((wordObj, i) => (
           <span
              key={wordObj.id}
              className={`px-3 py-1 rounded-full text-xs font-medium border ${
                i === 0 
                ? 'bg-slate-100 text-slate-600 border-slate-200' 
                : 'bg-slate-50 text-slate-400 border-slate-100 line-through'
              }`}
           >
             {wordObj.word}
           </span>
        ))}
        {sortedWords.length === 0 && (
          <p className="text-xs text-slate-400 italic p-2">No words called yet.</p>
        )}
      </div>
    </div>
  )
}
