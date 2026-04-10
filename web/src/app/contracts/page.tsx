'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Loader2, AlertTriangle, CheckCircle2, XCircle, Globe, Database, MessageSquare, Zap, Key, FileJson, Server } from 'lucide-react'
import Link from 'next/link'

interface Contract {
  id: string
  type: string
  role: string
  symbol_id: string
  file_path: string
  line: number
  repo_prefix?: string
  meta?: Record<string, string>
  confidence: number
}

interface MatchResult {
  matched: { contract_id: string; provider: Contract; consumer: Contract; cross_repo: boolean }[]
  orphan_providers: Contract[]
  orphan_consumers: Contract[]
}

const TYPE_ICONS: Record<string, React.ReactNode> = {
  http: <Globe className="h-4 w-4" />,
  grpc: <Server className="h-4 w-4" />,
  graphql: <Database className="h-4 w-4" />,
  topic: <MessageSquare className="h-4 w-4" />,
  ws: <Zap className="h-4 w-4" />,
  env: <Key className="h-4 w-4" />,
  openapi: <FileJson className="h-4 w-4" />,
}

const TYPE_COLORS: Record<string, string> = {
  http: 'bg-blue-500/20 text-blue-400',
  grpc: 'bg-purple-500/20 text-purple-400',
  graphql: 'bg-pink-500/20 text-pink-400',
  topic: 'bg-green-500/20 text-green-400',
  ws: 'bg-yellow-500/20 text-yellow-400',
  env: 'bg-orange-500/20 text-orange-400',
  openapi: 'bg-cyan-500/20 text-cyan-400',
}

export default function ContractsPage() {
  const [contracts, setContracts] = useState<Contract[]>([])
  const [matchResult, setMatchResult] = useState<MatchResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState('overview')

  useEffect(() => {
    async function load() {
      try {
        setLoading(true)
        const [contractsText, matchText] = await Promise.all([
          api.callTool('get_contracts', {}),
          api.callTool('check_contracts', {}),
        ])

        try {
          const parsed = JSON.parse(contractsText)
          setContracts(Array.isArray(parsed) ? parsed : parsed.contracts || [])
        } catch {
          setContracts([])
        }

        try {
          setMatchResult(JSON.parse(matchText))
        } catch {
          setMatchResult(null)
        }

        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load contracts')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-zinc-500" />
        <span className="ml-2 text-zinc-500">Loading contracts...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center">
        <Card className="max-w-md bg-red-950/30 border-red-800">
          <CardContent className="p-6 text-center">
            <AlertTriangle className="mx-auto h-8 w-8 text-red-400 mb-2" />
            <p className="text-red-300">{error}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Group contracts by type
  const byType: Record<string, Contract[]> = {}
  for (const c of contracts) {
    if (!byType[c.type]) byType[c.type] = []
    byType[c.type].push(c)
  }

  const providers = contracts.filter(c => c.role === 'provider')
  const consumers = contracts.filter(c => c.role === 'consumer')

  return (
    <div className="h-full overflow-auto p-6">
      <h1 className="text-2xl font-bold text-zinc-100 mb-6">API Contracts</h1>

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-zinc-100">{contracts.length}</div>
            <div className="text-xs text-zinc-500">Total Contracts</div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-green-400">{providers.length}</div>
            <div className="text-xs text-zinc-500">Providers</div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-blue-400">{consumers.length}</div>
            <div className="text-xs text-zinc-500">Consumers</div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900 border-zinc-800">
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-yellow-400">
              {(matchResult?.orphan_providers?.length || 0) + (matchResult?.orphan_consumers?.length || 0)}
            </div>
            <div className="text-xs text-zinc-500">Orphans</div>
          </CardContent>
        </Card>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="bg-zinc-900 border-zinc-800 mb-4">
          <TabsTrigger value="overview">By Type</TabsTrigger>
          <TabsTrigger value="matches">Matches</TabsTrigger>
          <TabsTrigger value="orphans">Orphans</TabsTrigger>
        </TabsList>

        {/* By Type */}
        <TabsContent value="overview">
          <div className="space-y-4">
            {Object.entries(byType).map(([type, items]) => (
              <Card key={type} className="bg-zinc-900 border-zinc-800">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-lg">
                    {TYPE_ICONS[type]}
                    <span className="capitalize">{type}</span>
                    <Badge variant="outline" className={TYPE_COLORS[type]}>{items.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-1">
                    {items.slice(0, 20).map((c, i) => (
                      <div key={i} className="flex items-center gap-2 text-sm py-1 border-b border-zinc-800 last:border-0">
                        <Badge variant="outline" className={c.role === 'provider' ? 'bg-green-500/20 text-green-400' : 'bg-blue-500/20 text-blue-400'}>
                          {c.role}
                        </Badge>
                        <code className="text-zinc-300 font-mono text-xs flex-1 truncate">{c.id}</code>
                        {c.symbol_id && (
                          <Link href={`/symbol/${encodeURIComponent(c.symbol_id)}`} className="text-xs text-zinc-500 hover:text-zinc-300 truncate max-w-48">
                            {c.symbol_id.split('::').pop()}
                          </Link>
                        )}
                        <span className="text-xs text-zinc-600">{c.file_path}</span>
                      </div>
                    ))}
                    {items.length > 20 && (
                      <div className="text-xs text-zinc-500 pt-1">...and {items.length - 20} more</div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
            {Object.keys(byType).length === 0 && (
              <div className="text-center text-zinc-500 py-12">No contracts detected</div>
            )}
          </div>
        </TabsContent>

        {/* Matches */}
        <TabsContent value="matches">
          <div className="space-y-2">
            {matchResult?.matched?.map((m, i) => (
              <Card key={i} className="bg-zinc-900 border-zinc-800">
                <CardContent className="p-4 flex items-center gap-3">
                  <CheckCircle2 className="h-4 w-4 text-green-400 shrink-0" />
                  <code className="text-xs font-mono text-zinc-300 flex-1">{m.contract_id}</code>
                  {m.cross_repo && <Badge variant="outline" className="bg-purple-500/20 text-purple-400">cross-repo</Badge>}
                  <div className="text-xs text-zinc-500">
                    <span className="text-green-400">{m.provider.file_path}</span>
                    {' → '}
                    <span className="text-blue-400">{m.consumer.file_path}</span>
                  </div>
                </CardContent>
              </Card>
            ))}
            {(!matchResult?.matched || matchResult.matched.length === 0) && (
              <div className="text-center text-zinc-500 py-12">No matched provider↔consumer pairs</div>
            )}
          </div>
        </TabsContent>

        {/* Orphans */}
        <TabsContent value="orphans">
          <div className="space-y-4">
            {matchResult?.orphan_providers && matchResult.orphan_providers.length > 0 && (
              <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-lg">
                    <XCircle className="h-4 w-4 text-yellow-400" />
                    Orphan Providers
                    <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400">{matchResult.orphan_providers.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-1">
                    {matchResult.orphan_providers.map((c, i) => (
                      <div key={i} className="flex items-center gap-2 text-sm py-1 border-b border-zinc-800 last:border-0">
                        <Badge variant="outline" className={TYPE_COLORS[c.type]}>{c.type}</Badge>
                        <code className="text-zinc-300 font-mono text-xs flex-1 truncate">{c.id}</code>
                        <span className="text-xs text-zinc-600">{c.file_path}:{c.line}</span>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {matchResult?.orphan_consumers && matchResult.orphan_consumers.length > 0 && (
              <Card className="bg-zinc-900 border-zinc-800">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-lg">
                    <XCircle className="h-4 w-4 text-red-400" />
                    Orphan Consumers
                    <Badge variant="outline" className="bg-red-500/20 text-red-400">{matchResult.orphan_consumers.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-1">
                    {matchResult.orphan_consumers.map((c, i) => (
                      <div key={i} className="flex items-center gap-2 text-sm py-1 border-b border-zinc-800 last:border-0">
                        <Badge variant="outline" className={TYPE_COLORS[c.type]}>{c.type}</Badge>
                        <code className="text-zinc-300 font-mono text-xs flex-1 truncate">{c.id}</code>
                        <span className="text-xs text-zinc-600">{c.file_path}:{c.line}</span>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {(!matchResult?.orphan_providers?.length && !matchResult?.orphan_consumers?.length) && (
              <div className="text-center text-zinc-500 py-12">No orphan contracts</div>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
