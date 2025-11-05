import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface VSTab {
    id: string
    label: string
    content?: React.ReactNode
    closeable?: boolean
}

interface VSTabsProps {
    tabs: VSTab[]
    activeTab: string
    onTabChange: (tabId: string) => void
    onTabClose?: (tabId: string) => void
    onTabMiddleClick?: (tabId: string) => void
}

export function VSTabs({ tabs, activeTab, onTabChange, onTabClose, onTabMiddleClick }: VSTabsProps) {
    const handleTabClick = (tabId: string, e: React.MouseEvent) => {
        // Middle click to close
        if (e.button === 1 && onTabMiddleClick) {
            e.preventDefault()
            e.stopPropagation()
            onTabMiddleClick(tabId)
        }
    }

    const handleCloseClick = (tabId: string, e: React.MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        if (onTabClose) {
            onTabClose(tabId)
        }
    }

    return (
        <Tabs value={activeTab} onValueChange={onTabChange} className='w-full'>
            <TabsList className='bg-background justify-start rounded-none border-b p-0 h-auto w-full overflow-x-auto'>
                {tabs.map(tab => (
                    <TabsTrigger
                        key={tab.id}
                        value={tab.id}
                        onMouseDown={(e) => handleTabClick(tab.id, e)}
                        className={cn(
                            'bg-background border-b-border dark:data-[state=active]:bg-background',
                            'data-[state=active]:border-border data-[state=active]:border-b-background',
                            'h-9 rounded-none rounded-t border border-transparent',
                            'data-[state=active]:-mb-0.5 data-[state=active]:shadow-none',
                            'dark:border-b-0 dark:data-[state=active]:-mb-0.5',
                            'px-3 py-1.5 group relative',
                            'hover:bg-accent/50 transition-colors',
                            'flex items-center gap-2 max-w-[200px]'
                        )}
                    >
                        <span className='truncate text-sm'>{tab.label}</span>
                        {tab.closeable !== false && onTabClose && (
                            <button
                                onClick={(e) => handleCloseClick(tab.id, e)}
                                className='ml-auto opacity-0 group-hover:opacity-100 hover:bg-accent rounded p-0.5 transition-opacity'
                                aria-label='Close tab'
                            >
                                <X className='h-3.5 w-3.5' />
                            </button>
                        )}
                    </TabsTrigger>
                ))}
            </TabsList>

            {tabs.map(tab => (
                tab.content && (
                    <TabsContent key={tab.id} value={tab.id} className='mt-0'>
                        {tab.content}
                    </TabsContent>
                )
            ))}
        </Tabs>
    )
}

// Demo component for reference
const tabs = [
    {
        name: 'Explore',
        value: 'explore',
        content: (
            <>
                Discover <span className='text-foreground font-semibold'>fresh ideas</span>, trending topics, and hidden gems
                curated just for you. Start exploring and let your curiosity lead the way!
            </>
        )
    },
    {
        name: 'Favorites',
        value: 'favorites',
        content: (
            <>
                All your <span className='text-foreground font-semibold'>favorites</span> are saved here. Revisit articles,
                collections, and moments you love, any time you want a little inspiration.
            </>
        )
    },
    {
        name: 'Surprise Me',
        value: 'surprise',
        content: (
            <>
                <span className='text-foreground font-semibold'>Surprise!</span> Here&apos;s something unexpected—a fun fact, a
                quirky tip, or a daily challenge. Come back for a new surprise every day!
            </>
        )
    }
]

const TabsLiftedDemo = () => {
    return (
        <div>
            <Tabs defaultValue='explore' className='gap-4'>
                <TabsList className='bg-background justify-start rounded-none border-b p-0'>
                    {tabs.map(tab => (
                        <TabsTrigger
                            key={tab.value}
                            value={tab.value}
                            className='bg-background border-b-border dark:data-[state=active]:bg-background data-[state=active]:border-border data-[state=active]:border-b-background h-full rounded-none rounded-t border border-transparent data-[state=active]:-mb-0.5 data-[state=active]:shadow-none dark:border-b-0 dark:data-[state=active]:-mb-0.5'
                        >
                            {tab.name}
                        </TabsTrigger>
                    ))}
                </TabsList>

                {tabs.map(tab => (
                    <TabsContent key={tab.value} value={tab.value}>
                        <p className='text-muted-foreground text-sm'>{tab.content}</p>
                    </TabsContent>
                ))}
            </Tabs>
        </div>
    )
}

export default TabsLiftedDemo
