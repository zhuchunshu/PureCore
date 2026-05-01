<script setup>
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import UserSecurity from '../../components/dashboard/UserSecurity.vue'

const { t } = useI18n()
useSEO({ title: t('user.security'), description: t('user.security') })
</script>

<template>
  <UserSecurity />
</template>
