'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '@/lib/api'
import type { Note } from '@/types'

interface Props {
  symbol: string
}

export default function NoteEditor({ symbol }: Props) {
  const qc = useQueryClient()
  const [content, setContent] = useState('')
  const [editing, setEditing] = useState<Note | null>(null)
  const [editContent, setEditContent] = useState('')

  const { data: notes = [] } = useQuery({
    queryKey: ['notes', symbol],
    queryFn: () => userApi.getNotes(symbol),
  })

  const create = useMutation({
    mutationFn: () => userApi.createNote(symbol, content),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notes', symbol] })
      setContent('')
    },
  })

  const update = useMutation({
    mutationFn: () => userApi.updateNote(editing!.id, editContent),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notes', symbol] })
      setEditing(null)
    },
  })

  const remove = useMutation({
    mutationFn: (id: string) => userApi.deleteNote(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notes', symbol] }),
  })

  return (
    <div className="mt-4">
      <h4 className="text-xs font-semibold text-text-secondary uppercase tracking-wider mb-3">
        📝 Ghi chú cá nhân — {symbol}
      </h4>

      {/* New note input */}
      <div className="flex gap-2 mb-3">
        <textarea
          value={content}
          onChange={e => setContent(e.target.value)}
          placeholder={`Ghi chú về ${symbol}...`}
          rows={2}
          className="flex-1 bg-bg-primary border border-bg-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:border-accent-blue resize-none transition-colors"
        />
        <button
          onClick={() => create.mutate()}
          disabled={!content.trim() || create.isPending}
          className="px-3 py-2 bg-accent-blue hover:bg-accent-blue/80 text-white text-xs font-semibold rounded-lg disabled:opacity-40 transition-all"
        >
          {create.isPending ? '...' : 'Lưu'}
        </button>
      </div>

      {/* Notes list */}
      <div className="space-y-2">
        {notes.map(note => (
          <div key={note.id} className="bg-bg-primary rounded-xl p-3 border border-bg-border group">
            {editing?.id === note.id ? (
              <div className="space-y-2">
                <textarea
                  value={editContent}
                  onChange={e => setEditContent(e.target.value)}
                  rows={3}
                  className="w-full bg-bg-card border border-accent-blue/40 rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none resize-none"
                />
                <div className="flex gap-2">
                  <button onClick={() => update.mutate()}
                    disabled={update.isPending}
                    className="text-xs bg-accent-blue text-white px-3 py-1.5 rounded-lg transition-all">
                    Cập nhật
                  </button>
                  <button onClick={() => setEditing(null)}
                    className="text-xs text-text-secondary hover:text-text-primary px-3 py-1.5 rounded-lg border border-bg-border">
                    Hủy
                  </button>
                </div>
              </div>
            ) : (
              <>
                <p className="text-sm text-text-primary whitespace-pre-wrap">{note.content}</p>
                <div className="flex items-center justify-between mt-2">
                  <span className="text-[10px] text-text-muted">
                    {new Date(note.updated_at).toLocaleString('vi-VN')}
                  </span>
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => { setEditing(note); setEditContent(note.content) }}
                      className="text-[10px] text-text-muted hover:text-accent-blue px-2 py-1 rounded transition-colors"
                    >✏️</button>
                    <button
                      onClick={() => remove.mutate(note.id)}
                      className="text-[10px] text-text-muted hover:text-accent-red px-2 py-1 rounded transition-colors"
                    >🗑</button>
                  </div>
                </div>
              </>
            )}
          </div>
        ))}

        {notes.length === 0 && (
          <p className="text-xs text-text-muted text-center py-3">
            Chưa có ghi chú nào cho {symbol}
          </p>
        )}
      </div>
    </div>
  )
}
