import { RealtimeCursors } from '@/components/realtime-cursors'

export default function Page() {
  return (
    <div className="w-full min-h-screen">
      <RealtimeCursors roomName="macrodata_refinement_office" />
    </div>
  )
}
