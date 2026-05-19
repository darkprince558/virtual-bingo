import { ActivityEvent } from "@/types/game"

interface ActivityFeedProps {
  events: ActivityEvent[]
}

const toneClass: Record<NonNullable<ActivityEvent["tone"]>, string> = {
  neutral: "bg-slate-400",
  success: "bg-emerald-500",
  warning: "bg-amber-500",
  danger: "bg-rose-500",
}

export function ActivityFeed({ events }: ActivityFeedProps) {
  return (
    <section className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
      <div className="px-5 py-4 border-b border-slate-200 bg-slate-50">
        <h3 className="text-[10px] font-bold uppercase tracking-[0.2em] text-slate-400">Activity Log</h3>
      </div>
      <ol className="divide-y divide-slate-100">
        {events.map((event) => (
          <li key={event.id} className="px-5 py-4 flex gap-3">
            <span className={`mt-1.5 h-2.5 w-2.5 rounded-full ${toneClass[event.tone ?? "neutral"]}`} />
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-sm font-bold text-slate-800">{event.label}</p>
                <time className="text-[10px] font-bold uppercase tracking-widest text-slate-400">
                  {new Date(event.createdAt).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}
                </time>
              </div>
              <p className="mt-1 text-xs font-medium leading-relaxed text-slate-500">{event.detail}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  )
}
