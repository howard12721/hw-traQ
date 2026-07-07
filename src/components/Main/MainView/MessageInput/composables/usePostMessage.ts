import { ref, unref } from 'vue'

import type { AxiosProgressEvent } from 'axios'

import useMessageInputStateStatic from '/@/composables/messageInputState/useMessageInputStateStatic'
import useChannelPath from '/@/composables/useChannelPath'
import apis, { buildFilePathForPost, formatResizeError } from '/@/lib/apis'
import { countLength } from '/@/lib/basic/string'
import { nullUuid } from '/@/lib/basic/uuid'
import { replace as embedInternalLink } from '/@/lib/markdown/internalLinkEmbedder'
import { isEmbeddedLink } from '/@/lib/markdown/markdown'
import { MESSAGE_MAX_LENGTH } from '/@/lib/validate'
import { useChannelsStore } from '/@/store/entities/channels'
import { useGroupsStore } from '/@/store/entities/groups'
import { useUsersStore } from '/@/store/entities/users'
import type {
  Attachment,
  MessageInputState,
  MessageInputStateKey
} from '/@/store/ui/messageInputStateStore'
import { useToastStore } from '/@/store/ui/toast'
import type { ChannelId } from '/@/types/entity-ids'

/**
 * @param progress アップロード進行状況 0～1
 */
type ProgressCallback = (progress: number) => void
const noopProgress: ProgressCallback = () => undefined

const uploadAttachments = async (
  attachments: ReadonlyArray<Readonly<Attachment>>,
  channelId: ChannelId,
  onProgress: ProgressCallback
) => {
  const responses = []
  for (const [i, attachment] of attachments.entries()) {
    responses.push(
      await apis.postFile(attachment.file, channelId, {
        /**
         * https://github.com/axios/axios#request-config
         */
        onUploadProgress(e: AxiosProgressEvent) {
          if (e.total === undefined || e.total === 0) return
          onProgress((i + e.loaded / e.total) / attachments.length)
        }
      })
    )
  }
  return responses.map(res => buildFilePathForPost(res.data.id))
}

export const createContent = async (
  embeddedText: string,
  fileUrls: string[]
) => {
  const joinContents = (delimiter: string, contents: string[]) =>
    contents.filter(Boolean).join(delimiter)

  const embeddedUrls = fileUrls.join('\n')
  const trimmedEmbeddedText = embeddedText.trimEnd()

  if (trimmedEmbeddedText === '') {
    return joinContents('\n', [embeddedText, embeddedUrls])
  }

  const trimmedEmbeddedTextLines = trimmedEmbeddedText.split(`\n`)

  if (
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    await isEmbeddedLink(trimmedEmbeddedTextLines.at(-1)!)
  ) {
    return joinContents('\n', [trimmedEmbeddedText, embeddedUrls])
  }

  return joinContents('\n\n', [embeddedText, embeddedUrls])
}

const usePostMessage = (
  channelId: MessageInputStateKey,
  inputStateKey = channelId
) => {
  const { getMessageInputState } = useMessageInputStateStatic()
  const { postMessageState } = usePostMessageSender()

  const isPosting = ref(false)
  const progress = ref(0)

  const postMessage = async () => {
    // awaitによって変化しないようにあえてリアクティブでないものを取得する
    const { state, isEmpty, clearState } = getMessageInputState(inputStateKey)
    // awaitの前でunrefしておかないと別のチャンネルに投稿されうる
    const cId = unref(channelId)

    if (isPosting.value || isEmpty) return false

    try {
      isPosting.value = true
      const posted = await postMessageState(cId, state, {
        onProgress: p => {
          progress.value = p
        }
      })

      if (posted) {
        clearState()
      }
      return posted
    } finally {
      isPosting.value = false
      progress.value = 0
    }
  }
  return { postMessage, isPosting, progress }
}

interface PostMessageStateOptions {
  skipForceConfirm?: boolean
  onProgress?: ProgressCallback
  errorMessage?: string
}

interface PrepareMessageInputContentOptions {
  onProgress?: ProgressCallback
}

export const usePostMessageSender = () => {
  const { channelPathStringToId, channelIdToShortPathString } = useChannelPath()
  const { addErrorToast } = useToastStore()
  const { channelsMap, fetchChannels } = useChannelsStore()
  const { findUserByName, fetchUsers } = useUsersStore()
  const { getUserGroupByName, fetchUserGroups } = useGroupsStore()

  const getForceConfirmString = (cId: ChannelId) =>
    `${channelIdToShortPathString(
      cId,
      true
    )}に投稿されたメッセージは全員に通知されます。メッセージを投稿しますか？\n注) このチャンネルは重要な連絡以外には使用しないでください。`

  const confirmForcePostIfNeeded = (
    cId: ChannelId,
    skipForceConfirm = false
  ) => {
    if (skipForceConfirm) return true
    if (!channelsMap.value.get(cId)?.force) return true

    return confirm(getForceConfirmString(cId))
  }

  const prepareContent = async (state: Readonly<MessageInputState>) => {
    await Promise.all([fetchUsers(), fetchUserGroups(), fetchChannels()])

    const embeddedText = embedInternalLink(state.text, {
      getUser: findUserByName,
      getGroup: getUserGroupByName,
      getChannel: path => {
        try {
          const id = channelPathStringToId(path)
          return { id }
        } catch {
          return undefined
        }
      }
    })

    const dummyFileUrls = state.attachments.map(() =>
      buildFilePathForPost(nullUuid)
    )
    const dummyText = await createContent(embeddedText, dummyFileUrls)
    if (countLength(dummyText) > MESSAGE_MAX_LENGTH) {
      addErrorToast('メッセージが長すぎます')
      return undefined
    }

    return embeddedText
  }

  const validateMessageInputState = async (
    state: Readonly<MessageInputState>
  ) => (await prepareContent(state)) !== undefined

  const prepareMessageInputContent = async (
    cId: ChannelId,
    state: Readonly<MessageInputState>,
    { onProgress = noopProgress }: PrepareMessageInputContentOptions = {}
  ) => {
    const embeddedText = await prepareContent(state)
    if (embeddedText === undefined) return undefined

    const fileUrls = await uploadAttachments(state.attachments, cId, onProgress)
    return createContent(embeddedText, fileUrls)
  }

  const postMessageState = async (
    cId: ChannelId,
    state: Readonly<MessageInputState>,
    {
      skipForceConfirm = false,
      onProgress = noopProgress,
      errorMessage = 'メッセージ送信に失敗しました'
    }: PostMessageStateOptions = {}
  ) => {
    if (!confirmForcePostIfNeeded(cId, skipForceConfirm)) {
      return false
    }

    try {
      const content = await prepareMessageInputContent(cId, state, {
        onProgress
      })
      if (content === undefined) return false

      await apis.postMessage(cId, { content })

      return true
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error('メッセージ送信に失敗しました', e)

      addErrorToast(formatResizeError(e, errorMessage))
      return false
    }
  }

  return {
    postMessageState,
    validateMessageInputState,
    prepareMessageInputContent,
    confirmForcePostIfNeeded,
    getForceConfirmString
  }
}

export default usePostMessage
