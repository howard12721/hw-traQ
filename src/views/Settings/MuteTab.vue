<template>
  <section :class="$style.container">
    <section :class="$style.section">
      <div :class="$style.description">
        <h3>ユーザー</h3>
        <p>指定したユーザーのメッセージ本文を隠します。</p>
      </div>
      <div :class="$style.addRow">
        <FilterInput
          v-model="userQuery"
          :class="$style.input"
          placeholder="ユーザーを検索"
          enterkeyhint="done"
          @enter="addFirstUser"
        />
        <FormButton
          label="追加"
          icon="plus"
          mdi
          type="secondary"
          :disabled="!firstUser"
          :class="$style.addButton"
          @click="addFirstUser"
        />
      </div>
      <p v-if="firstUser" :class="$style.hint">
        追加候補: {{ firstUser.displayName }} @{{ firstUser.name }}
      </p>
      <ul v-if="mutedUserItems.length > 0" :class="$style.list">
        <li
          v-for="item in mutedUserItems"
          :key="item.id"
          :class="$style.userItem"
        >
          <UserIcon :user-id="item.id" :size="28" prevent-modal />
          <span :class="$style.userName">
            {{ item.user?.displayName ?? item.id }}
            <span v-if="item.user" :class="$style.screenName">
              @{{ item.user.name }}
            </span>
          </span>
          <button
            type="button"
            :class="$style.removeButton"
            title="ミュートを解除"
            @click="removeMutedUserId(item.id)"
          >
            <AIcon mdi name="close" :size="18" />
          </button>
        </li>
      </ul>
      <p v-else :class="$style.empty">ミュート中のユーザーはいません。</p>
    </section>

    <section :class="$style.section">
      <div :class="$style.description">
        <h3>キーワード</h3>
        <p>本文に指定したキーワードを含むメッセージを隠します。</p>
      </div>
      <div :class="$style.addRow">
        <FilterInput
          v-model="keywordInput"
          :class="$style.input"
          placeholder="キーワードを入力"
          enterkeyhint="done"
          @enter="addKeyword"
        />
        <FormButton
          label="追加"
          icon="plus"
          mdi
          type="secondary"
          :disabled="keywordInput.trim() === ''"
          :class="$style.addButton"
          @click="addKeyword"
        />
      </div>
      <ul v-if="normalizedMutedKeywords.length > 0" :class="$style.keywordList">
        <li
          v-for="keyword in normalizedMutedKeywords"
          :key="keyword"
          :class="$style.keywordItem"
        >
          <span :class="$style.keyword">{{ keyword }}</span>
          <button
            type="button"
            :class="$style.removeButton"
            title="ミュートを解除"
            @click="removeMutedKeyword(keyword)"
          >
            <AIcon mdi name="close" :size="18" />
          </button>
        </li>
      </ul>
      <p v-else :class="$style.empty">ミュート中のキーワードはありません。</p>
    </section>
  </section>
</template>

<script lang="ts" setup>
import type { User } from '@traptitech/traq'

import { computed, onMounted, ref } from 'vue'

import AIcon from '/@/components/UI/AIcon.vue'
import FilterInput from '/@/components/UI/FilterInput.vue'
import FormButton from '/@/components/UI/FormButton.vue'
import UserIcon from '/@/components/UI/UserIcon.vue'
import useUserList from '/@/composables/users/useUserList'
import useTextFilter from '/@/composables/utils/useTextFilter'
import { useMuteSettings } from '/@/store/app/muteSettings'
import { useUsersStore } from '/@/store/entities/users'

const {
  mutedUserIds,
  normalizedMutedKeywords,
  addMutedUserId,
  removeMutedUserId,
  addMutedKeyword,
  removeMutedKeyword
} = useMuteSettings()
const { usersMap, fetchUsers } = useUsersStore()

onMounted(() => {
  void fetchUsers({})
})

const userList = useUserList(computed(() => ['inactive', 'webhook']))
const { query: userQuery, filteredItems: filteredUsers } = useTextFilter(
  userList,
  ['name', 'displayName'],
  { limit: 20 }
)
const addableUsers = computed(() =>
  userQuery.value.trim() === ''
    ? []
    : filteredUsers.value.filter(
        user => !mutedUserIds.value.includes(user.id)
      )
)
const firstUser = computed(() => addableUsers.value[0])
const mutedUserItems = computed(() =>
  mutedUserIds.value.map(id => ({
    id,
    user: usersMap.value.get(id) as User | undefined
  }))
)

const addFirstUser = () => {
  if (!firstUser.value) return
  addMutedUserId(firstUser.value.id)
  userQuery.value = ''
}

const keywordInput = ref('')
const addKeyword = () => {
  addMutedKeyword(keywordInput.value)
  keywordInput.value = ''
}
</script>

<style lang="scss" module>
.container {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.description {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.addRow {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input {
  flex: 1 1 240px;
  min-width: 0;
}

.addButton {
  flex-shrink: 0;
}

.list,
.keywordList {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.userItem,
.keywordItem {
  @include background-secondary;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  padding: 8px;
  border-radius: 4px;
}

.userName,
.keyword {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.userName {
  flex: 1;
}

.screenName,
.empty,
.hint {
  @include color-ui-secondary;
}

.keyword {
  flex: 1;
}

.removeButton {
  @include color-ui-secondary;
  display: flex;
  flex-shrink: 0;
  padding: 4px;
  border-radius: 4px;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    @include color-ui-primary;
    @include background-tertiary;
  }
}

@media (max-width: 600px) {
  .addRow {
    align-items: stretch;
    flex-direction: column;
  }

  .addButton {
    align-self: flex-start;
  }
}
</style>
