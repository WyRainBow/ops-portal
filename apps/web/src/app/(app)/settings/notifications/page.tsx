'use client'

import { useCallback, useEffect, useState } from 'react'
import { getToken } from '../../../../lib/auth'
import { Badge, Button, Card, Input } from '../../../../components/Ui'

interface NotificationConfig {
  feishu_webhook: string
  feishu_enabled: boolean
  alert_severity: string[]
  notify_on_resolved: boolean
}

export default function NotificationsPage() {
  const token = getToken() || ''

  const [config, setConfig] = useState<NotificationConfig>({
    feishu_webhook: '',
    feishu_enabled: false,
    alert_severity: ['critical', 'high'],
    notify_on_resolved: true,
  })
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const loadConfig = useCallback(async () => {
    if (!token) return
    setLoading(true)
    setErr(null)

    // In production, this would fetch from API
    // Simulating with localStorage for demo
    const saved = localStorage.getItem('notification_config')
    if (saved) {
      setConfig(JSON.parse(saved))
    }

    setLoading(false)
  }, [token])

  useEffect(() => {
    void loadConfig()
  }, [loadConfig])

  const handleSave = useCallback(async () => {
    if (!token) return

    setSaving(true)
    setErr(null)
    setSuccess(null)

    try {
      // In production, this would call API
      localStorage.setItem('notification_config', JSON.stringify(config))
      setSuccess('配置已保存')
      setTimeout(() => setSuccess(null), 3000)
    } catch (e: any) {
      setErr(e?.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }, [config, token])

  const handleTest = useCallback(async () => {
    if (!token) return

    setTesting(true)
    setErr(null)
    setSuccess(null)

    try {
      // In production, this would call test API
      await new Promise(resolve => setTimeout(resolve, 1000))
      setSuccess('测试通知已发送（模拟）')
      setTimeout(() => setSuccess(null), 3000)
    } catch (e: any) {
      setErr(e?.message || '发送失败')
    } finally {
      setTesting(false)
    }
  }, [token])

  const toggleSeverity = (severity: string) => {
    setConfig(prev => ({
      ...prev,
      alert_severity: prev.alert_severity.includes(severity)
        ? prev.alert_severity.filter(s => s !== severity)
        : [...prev.alert_severity, severity],
    }))
  }

  const severityLevels = ['critical', 'high', 'warning', 'info']

  return (
    <div className="space-y-6">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-slate-300/70">Settings</div>
        <div className="mt-1 text-2xl font-semibold">通知配置</div>
        <div className="mt-2 text-sm text-slate-200/70">配置飞书等通知渠道，接收告警推送</div>
      </div>

      {err && (
        <Card right={<Badge tone="bad">error</Badge>}>
          <div className="text-rose-400">{err}</div>
        </Card>
      )}

      {success && (
        <Card right={<Badge tone="ok">success</Badge>}>
          <div className="text-emerald-400">{success}</div>
        </Card>
      )}

      {/* Feishu Configuration */}
      <Card title="飞书通知" right={loading ? <Badge tone="warn">loading</Badge> : null}>
        <div className="space-y-4">
          <div>
            <div className="mb-1 text-xs text-slate-200/60">启用飞书通知</div>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.feishu_enabled}
                onChange={(e) => setConfig({ ...config, feishu_enabled: e.target.checked })}
                className="h-4 w-4 rounded"
              />
              <span className="text-sm">启用飞书机器人推送</span>
            </label>
          </div>

          <div>
            <div className="mb-1 text-xs text-slate-200/60">Webhook URL</div>
            <Input
              value={config.feishu_webhook}
              onChange={(e) => setConfig({ ...config, feishu_webhook: e.target.value })}
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..."
              className="font-mono text-sm"
            />
            <div className="mt-1 text-xs text-slate-200/50">
              在飞书群设置中添加机器人获取 Webhook 地址
            </div>
          </div>

          <div>
            <div className="mb-2 text-xs text-slate-200/60">告警级别（多选）</div>
            <div className="flex flex-wrap gap-2">
              {severityLevels.map(level => (
                <button
                  key={level}
                  onClick={() => toggleSeverity(level)}
                  className={`rounded-full border px-3 py-1.5 text-xs transition ${
                    config.alert_severity.includes(level)
                      ? 'border-slate-900 bg-slate-900 text-white'
                      : 'border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
                  }`}
                >
                  {level}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={config.notify_on_resolved}
                onChange={(e) => setConfig({ ...config, notify_on_resolved: e.target.checked })}
                className="h-4 w-4 rounded"
              />
              <span className="text-sm">告警恢复时也发送通知</span>
            </label>
          </div>
        </div>
      </Card>

      {/* Preview */}
      <Card title="通知预览" right={<Badge tone="neutral">preview</Badge>}>
        <div className="rounded-xl border border-white/10 bg-slate-800 p-4">
          <div className="mb-3 flex items-center gap-2">
            <div className="h-2 w-2 rounded-full bg-rose-500" />
            <span className="font-semibold text-sm">【Critical 告警】HighCPUUsage</span>
          </div>
          <div className="space-y-2 text-xs text-slate-300">
            <div>状态: firing</div>
            <div>级别: critical</div>
            <div>主机: server-01</div>
            <div>时间: {new Date().toLocaleString('zh-CN')}</div>
            <div className="mt-3 text-slate-400">摘要: CPU 使用率超过 90%</div>
          </div>
        </div>
      </Card>

      {/* Actions */}
      <div className="flex gap-3">
        <Button
          tone="ghost"
          onClick={handleTest}
          disabled={testing || !config.feishu_enabled}
        >
          {testing ? '发送中...' : '发送测试'}
        </Button>
        <Button
          tone="primary"
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? '保存中...' : '保存配置'}
        </Button>
      </div>

      {/* Help */}
      <Card title="配置说明" right={<Badge tone="neutral">Help</Badge>}>
        <div className="space-y-2 text-sm text-slate-200/70">
          <p><strong>1. 创建飞书机器人：</strong></p>
          <p className="text-slate-200/50">在飞书群 → 群设置 → 群机器人 → 添加机器人 → 自定义机器人</p>
          <p className="mt-2"><strong>2. 获取 Webhook URL：</strong></p>
          <p className="text-slate-200/50">创建机器人后会生成 Webhook 地址，复制粘贴到上方输入框</p>
          <p className="mt-2"><strong>3. 配置告警规则：</strong></p>
          <p className="text-slate-200/50">选择需要推送的告警级别，建议至少启用 critical 和 high</p>
        </div>
      </Card>
    </div>
  )
}
