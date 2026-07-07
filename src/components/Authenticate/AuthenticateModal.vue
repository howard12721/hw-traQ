<template>
  <div :class="$style.container">
    <slot />
  </div>
</template>

<style lang="scss" module>
.container {
  @include background-primary;
  border: 2px solid $theme-background-primary-border;
  border-radius: 4px;
  box-shadow: 8px 8px 0 $theme-background-secondary-border;
  width: 100%;
  max-width: 480px;
  padding: 48px;
  margin: auto;
}
</style>
