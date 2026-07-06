<template>
  <section :class="$style.container">
    <section :class="$style.section">
      <h3 :class="$style.heading">エントリー</h3>
      <div
        v-for="(entry, index) in state.entries"
        :key="entry.key"
        :class="$style.entry"
      >
        <div :class="$style.entryHeader">
          <span :class="$style.entryTitle">#{{ index + 1 }}</span>
          <FormButton
            label="削除"
            type="tertiary"
            :disabled="saving"
            @click="removeEntry(index)"
          />
        </div>
        <FormInput
          v-model="entry.pattern"
          placeholder="例: release|障害|deploy"
          autocomplete="off"
        />
        <div :class="$style.options">
          <FormCheckbox v-model="entry.includeSelf">
            自分自身の投稿を含める
          </FormCheckbox>
          <FormCheckbox v-model="entry.includeBots">
            BOTの投稿を含める
          </FormCheckbox>
        </div>
      </div>
      <div :class="$style.addEntry">
        <FormButton label="追加" type="tertiary" @click="addEntry" />
      </div>
    </section>

    <section :class="$style.section">
      <h3 :class="$style.heading">状態</h3>
      <p :class="$style.status" :data-running="$boolAttr(status.running)">
        {{ statusLabel }}
      </p>
      <p :class="$style.caption">{{ tokenLabel }}</p>
      <p v-if="errorMessage" :class="$style.error">{{ errorMessage }}</p>
    </section>

    <div :class="$style.buttons">
      <FormButton
        label="アクセストークンを発行"
        type="tertiary"
        :loading="issuingToken"
        :disabled="loading || saving"
        @click="issueToken"
      />
      <FormButton
        label="再読み込み"
        type="tertiary"
        :disabled="loading || saving || issuingToken"
        @click="load"
      />
      <FormButton
        label="保存"
        :loading="saving"
        :disabled="loading || issuingToken"
        @click="save"
      />
    </div>
  </section>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue'

import FormButton from '/@/components/UI/FormButton.vue'
import FormCheckbox from '/@/components/UI/FormCheckbox.vue'
import FormInput from '/@/components/UI/FormInput.vue'
import {
  consumeGazerTokenCallback,
  startGazerTokenIssue
} from '/@/lib/gazerOAuth'
import type { GazerEntry } from '/@/lib/internalApi'
import { getGazer, putGazer, putGazerToken } from '/@/lib/internalApi'
import { useGazerNotificationsStore } from '/@/store/domain/gazerNotifications'
import { useToastStore } from '/@/store/ui/toast'

type GazerEntryState = GazerEntry & {
  key: number
}

let nextEntryKey = 1
const createEntry = (entry?: GazerEntry): GazerEntryState => ({
  key: nextEntryKey++,
  pattern: entry?.pattern ?? '',
  includeSelf: entry?.includeSelf ?? false,
  includeBots: entry?.includeBots ?? false
})

const state = reactive({
  entries: [createEntry()] as GazerEntryState[]
})
const status = reactive({
  running: false,
  enabled: false,
  tokenConfigured: false
})
const loading = ref(false)
const saving = ref(false)
const issuingToken = ref(false)
const errorMessage = ref('')

const { addSuccessToast, addErrorToast } = useToastStore()
const { applyGazerResponse } = useGazerNotificationsStore()

const statusLabel = computed(() => {
  if (status.running) return '監視中'
  if (status.enabled) return '設定済み'
  return '停止中'
})
const tokenLabel = computed(() =>
  status.tokenConfigured ? 'アクセストークン設定済み' : 'アクセストークン未設定'
)

const applyResponse = (res: Awaited<ReturnType<typeof getGazer>>) => {
  applyGazerResponse(res)
  const entries = res.setting.entries.map(entry => createEntry(entry))
  state.entries = entries.length > 0 ? entries : [createEntry()]
  status.enabled = res.setting.enabled
  status.running = res.status.running
  status.tokenConfigured = res.status.tokenConfigured
}

const addEntry = () => {
  state.entries.push(createEntry())
}

const removeEntry = (index: number) => {
  if (state.entries.length <= 1) {
    state.entries = [createEntry()]
    return
  }
  state.entries.splice(index, 1)
}

const load = async () => {
  try {
    loading.value = true
    errorMessage.value = ''
    applyResponse(await getGazer())
  } catch (e) {
    errorMessage.value = 'Gazer設定を取得できませんでした'
    addErrorToast(errorMessage.value)
    // eslint-disable-next-line no-console
    console.error(e)
  } finally {
    loading.value = false
  }
}

const save = async () => {
  try {
    saving.value = true
    errorMessage.value = ''
    applyResponse(
      await putGazer({
        entries: state.entries.map(entry => ({
          pattern: entry.pattern,
          includeSelf: entry.includeSelf,
          includeBots: entry.includeBots
        }))
      })
    )
    addSuccessToast('Gazer設定を保存しました')
  } catch (e) {
    errorMessage.value = '正規表現または保存内容が不正です'
    addErrorToast(errorMessage.value)
    // eslint-disable-next-line no-console
    console.error(e)
  } finally {
    saving.value = false
  }
}

const handleTokenCallback = async () => {
  try {
    const callback = consumeGazerTokenCallback()
    if (!callback) return
    applyResponse(await putGazerToken(callback))
    addSuccessToast('Gazer用アクセストークンを保存しました')
  } catch (e) {
    errorMessage.value = 'Gazer用アクセストークンを保存できませんでした'
    addErrorToast(errorMessage.value)
    // eslint-disable-next-line no-console
    console.error(e)
  }
}

const issueToken = async () => {
  try {
    issuingToken.value = true
    errorMessage.value = ''
    await startGazerTokenIssue()
  } catch (e) {
    issuingToken.value = false
    errorMessage.value = 'Gazer用アクセストークンを発行できませんでした'
    addErrorToast(errorMessage.value)
    // eslint-disable-next-line no-console
    console.error(e)
  }
}

onMounted(async () => {
  try {
    await handleTokenCallback()
  } finally {
    await load()
  }
})
</script>

<style lang="scss" module>
.container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.entryHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.heading {
  @include color-ui-primary;
  @include size-body1;
  font-weight: bold;
}

.entry {
  @include background-secondary;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
  border-radius: 4px;
}

.entryTitle {
  @include color-ui-secondary;
  font-weight: bold;
}

.caption {
  @include color-ui-secondary;
  @include size-caption;
}

.options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.addEntry {
  display: flex;
  justify-content: flex-start;
}

.status {
  @include color-ui-secondary;
  &[data-running] {
    color: $theme-accent-primary-default;
    font-weight: bold;
  }
}

.error {
  color: $theme-accent-error-default;
  font-weight: bold;
}

.buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
