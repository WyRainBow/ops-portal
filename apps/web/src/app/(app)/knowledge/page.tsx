'use client'

import { useCallback, useEffect, useState } from 'react'
import { getToken } from '../../../lib/auth'
import { uploadDocument, chat } from '../../../lib/api'
import { Badge, Button, Card, Input, TextArea } from '../../../components/Ui'

interface Document {
  id: string
  source: string
  content: string
  metadata: Record<string, any>
}

export default function KnowledgePage() {
  const token = getToken() || ''

  const [documents, setDocuments] = useState<Document[]>([])
  const [loading, setLoading] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  // Upload state
  const [file, setFile] = useState<File | null>(null)
  const [uploadStatus, setUploadStatus] = useState<string>('')

  // Search state
  const [query, setQuery] = useState('')
  const [searching, setSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<string>('')

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0]
    if (selected) {
      setFile(selected)
      setUploadStatus('')
    }
  }, [])

  const handleUpload = useCallback(async () => {
    if (!file || !token) return

    setUploading(true)
    setErr(null)
    setUploadStatus('')

    try {
      const result = await uploadDocument(token, file)
      setUploadStatus(`✓ 上传成功: ${result.fileName}`)
      setFile(null)
      // Reset file input
      const fileInput = document.getElementById('file-upload') as HTMLInputElement
      if (fileInput) fileInput.value = ''
    } catch (e: any) {
      setErr(e?.message || '上传失败')
      setUploadStatus(`✗ 上传失败: ${e?.message}`)
    } finally {
      setUploading(false)
    }
  }, [file, token])

  const handleSearch = useCallback(async () => {
    if (!query.trim() || !token) return

    setSearching(true)
    setErr(null)
    setSearchResults('')

    try {
      const result = await chat(token, { question: query })
      setSearchResults(result.result || '无相关结果')
    } catch (e: any) {
      setErr(e?.message || '搜索失败')
    } finally {
      setSearching(false)
    }
  }, [query, token])

  // Example documents for demo
  const exampleDocs: Document[] = [
    {
      id: '1',
      source: '运维手册.md',
      content: '服务器重启流程...',
      metadata: { _source: 'ops-manual.md', _type: 'markdown' },
    },
    {
      id: '2',
      source: '故障排查指南.md',
      content: '常见故障处理方法...',
      metadata: { _source: 'troubleshooting.md', _type: 'markdown' },
    },
    {
      id: '3',
      source: 'API文档.md',
      content: 'API接口说明...',
      metadata: { _source: 'api-docs.md', _type: 'markdown' },
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Knowledge</div>
        <div className="mt-1 text-2xl font-semibold">知识库管理</div>
        <div className="mt-2 text-sm text-slate-200/70">上传文档、构建向量索引、测试 RAG 检索</div>
      </div>

      {/* Upload Section */}
      <Card title="文档上传" subtitle="支持 Markdown、TXT、PDF 等格式" right={err ? <Badge tone="bad">{err}</Badge> : null}>
        <div className="space-y-4">
          <div className="flex items-center gap-4">
            <input
              id="file-upload"
              type="file"
              accept=".md,.txt,.pdf,.doc,.docx"
              onChange={handleFileSelect}
              className="flex-1 rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-sm file:mr-4 file:rounded-2xl file:border-0 file:bg-white/10 file:px-4 file:py-2"
            />
            <Button
              tone="primary"
              onClick={handleUpload}
              disabled={!file || uploading}
            >
              {uploading ? '上传中...' : '上传'}
            </Button>
          </div>
          {file && (
            <div className="text-sm text-slate-200/70">
              已选择: {file.name} ({(file.size / 1024).toFixed(2)} KB)
            </div>
          )}
          {uploadStatus && (
            <div className={`text-sm ${uploadStatus.startsWith('✓') ? 'text-emerald-400' : 'text-rose-400'}`}>
              {uploadStatus}
            </div>
          )}
          <div className="text-xs text-slate-200/50">
            上传后会自动进行分块、向量化并存储到 Milvus
          </div>
        </div>
      </Card>

      {/* Search Section */}
      <Card title="向量搜索测试" subtitle="测试 RAG 检索效果">
        <div className="space-y-4">
          <div className="flex gap-3">
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="输入问题测试检索效果..."
              onKeyDown={(e) => e.key === 'Enter' && void handleSearch()}
              className="flex-1"
            />
            <Button
              tone="primary"
              onClick={handleSearch}
              disabled={!query.trim() || searching}
            >
              {searching ? '搜索中...' : '搜索'}
            </Button>
          </div>
          {searchResults && (
            <div>
              <div className="mb-2 text-xs text-slate-200/60">搜索结果</div>
              <div className="rounded-xl bg-black/30 p-4 text-sm text-slate-300">
                <pre className="whitespace-pre-wrap">{searchResults}</pre>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Documents List */}
      <Card title="已索引文档" subtitle="示例数据（实际应从 Milvus 获取）" right={<Badge tone="neutral">{exampleDocs.length} 个</Badge>}>
        <div className="overflow-hidden rounded-2xl border border-white/10">
          <table className="w-full text-left text-sm">
            <thead className="bg-white/5 text-xs font-semibold uppercase tracking-wide text-slate-200/70">
              <tr>
                <th className="px-4 py-3">ID</th>
                <th className="px-4 py-3">来源</th>
                <th className="px-4 py-3">类型</th>
                <th className="px-4 py-3">状态</th>
                <th className="px-4 py-3">操作</th>
              </tr>
            </thead>
            <tbody>
              {exampleDocs.map((doc) => (
                <tr key={doc.id} className="border-t border-white/5 hover:bg-white/5">
                  <td className="px-4 py-3 font-mono">{doc.id}</td>
                  <td className="px-4 py-3">{doc.source}</td>
                  <td className="px-4 py-3">{doc.metadata._type}</td>
                  <td className="px-4 py-3">
                    <Badge tone="ok">已索引</Badge>
                  </td>
                  <td className="px-4 py-3">
                    <Button tone="ghost" className="text-xs">
                      删除
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Info Card */}
      <Card title="知识库说明" right={<Badge tone="neutral">Info</Badge>}>
        <div className="space-y-3 text-sm text-slate-200/70">
          <p><strong>工作流程：</strong></p>
          <ol className="list-decimal list-inside space-y-1 text-slate-200/60">
            <li>上传文档（支持 Markdown、TXT、PDF 等格式）</li>
            <li>文档分块（按段落或语义分割）</li>
            <li>向量化（使用 Doubao Embedding 模型）</li>
            <li>存储到 Milvus 向量数据库</li>
            <li>支持语义搜索和混合检索</li>
          </ol>
          <p className="mt-4"><strong>重排序：</strong></p>
          <p className="text-slate-200/60">召回后使用 Cross-Encoder 模型进行重排序，提高检索精度</p>
        </div>
      </Card>
    </div>
  )
}
