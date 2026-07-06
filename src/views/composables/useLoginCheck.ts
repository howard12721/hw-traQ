import { onActivated, onBeforeMount, ref } from 'vue'

import { createSingleflight } from '/@/lib/basic/async'
import { getGazer } from '/@/lib/internalApi'
import { setupWebSocket } from '/@/lib/websocket'
import router, { RouteName } from '/@/router'
import { useGazerNotificationsStore } from '/@/store/domain/gazerNotifications'
import { useMeStore } from '/@/store/domain/me'

/**
 * ログイン状態かを確認し、ログインしていなかった場合はログイン画面へ遷移する
 */
const performLoginCheck = createSingleflight(
  async (fetchMe: () => Promise<object | undefined>) => {
    const res = await fetchMe()
    if (!res) {
      router.replace({
        name: RouteName.Login,
        query: { redirect: `${location.pathname}${location.search}` }
      })
      throw new Error('Login required')
    }
  }
)

const syncGazerSession = createSingleflight(async () => {
  try {
    const res = await getGazer()
    const gazerNotificationsStore = useGazerNotificationsStore()
    gazerNotificationsStore.applyGazerResponse(res)
    await gazerNotificationsStore.fetchNotifications()
  } catch (e) {
    // eslint-disable-next-line no-console
    console.warn('Failed to sync gazer session', { cause: e })
  }
})

/**
 * @param afterCheck ログイン確認後、ログインしていたら実行される
 */
const useLoginCheck = (afterCheck?: () => void) => {
  const { detail, fetchMe } = useMeStore()
  const isLoginCheckDone = ref(false)

  const hook = async () => {
    // 不整合の防止のため常にリクエストを送る
    try {
      await performLoginCheck(fetchMe)
    } catch {}

    if (detail.value !== undefined) {
      await setupWebSocket()
      void syncGazerSession()

      afterCheck?.()
    }

    isLoginCheckDone.value = true
  }
  onBeforeMount(hook)
  onActivated(hook)

  return { isLoginCheckDone }
}

export default useLoginCheck
