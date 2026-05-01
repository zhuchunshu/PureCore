<script setup>
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import UserSessions from '../../components/dashboard/UserSessions.vue'

const { t } = useI18n()
useSEO({ title: t('user.sessions'), description: t('user.sessions_description') })
</script>

<template>
  <UserSessions />
</template>
