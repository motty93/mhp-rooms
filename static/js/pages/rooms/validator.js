export function validateCreateForm(form) {
  return !!(form?.name?.trim() && form?.gameVersionId && form?.maxPlayers)
}
