<script setup>
import { useI18n } from '../../i18n'
import { useSEO } from '../../composables/useSEO'
import UserApiKeys from '../../components/dashboard/UserApiKeys.vue'

const { t } = useI18n()
useSEO({ title: t('user.api_keys'), description: t('user.api_keys_description') })
</script>

<template>
  <UserApiKeys />
</template>
