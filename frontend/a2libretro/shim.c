#include <string.h>

#include "shim.h"
#include "_cgo_export.h"

/*
These two are answered in C, without entering Go.

Every function exported from Go begins by waiting for the Go runtime to finish
starting, and the frontend asks a core what it is on its main thread, on the way
to loading it, which is the same thread the runtime needs to finish starting on.
Answering from Go there deadlocks the frontend for good: RetroArch sits in
_cgo_wait_runtime_init_done inside libretro_get_system_info and never returns.

They are constants anyway, so there is nothing to ask Go for. Keep them in step
with izapple2_libretro.info.
*/

unsigned retro_api_version(void)
{
   return RETRO_API_VERSION;
}

void retro_get_system_info(struct retro_system_info *info)
{
   memset(info, 0, sizeof(*info));
   info->library_name     = "izapple2";
   info->library_version  = "2.0";
   info->valid_extensions = "dsk|do|po|nib|woz|2mg|hdv|wav|zip|gz|m3u";
   /* The core opens the disks by name, it is not given their contents */
   info->need_fullpath    = true;
   /* The frontend unpacks the archives it knows, the core reads the rest */
   info->block_extract    = false;
}

static retro_log_printf_t log_cb;

void shim_set_log(retro_log_printf_t cb) { log_cb = cb; }
bool shim_has_log(void) { return log_cb != NULL; }

void shim_log(int level, const char *message)
{
   if (log_cb)
      log_cb((enum retro_log_level)level, "%s\n", message);
}

static retro_environment_t env_cb;
static retro_video_refresh_t video_cb;
static retro_audio_sample_batch_t audio_batch_cb;
static retro_input_poll_t input_poll_cb;
static retro_input_state_t input_state_cb;

void shim_set_environment(retro_environment_t cb) { env_cb = cb; }
void shim_set_video_refresh(retro_video_refresh_t cb) { video_cb = cb; }
void shim_set_audio_sample_batch(retro_audio_sample_batch_t cb) { audio_batch_cb = cb; }
void shim_set_input_poll(retro_input_poll_t cb) { input_poll_cb = cb; }
void shim_set_input_state(retro_input_state_t cb) { input_state_cb = cb; }

bool shim_environment(unsigned cmd, void *data)
{
   if (!env_cb)
      return false;
   return env_cb(cmd, data);
}

void shim_video_refresh(const void *data, unsigned width, unsigned height, size_t pitch)
{
   if (video_cb)
      video_cb(data, width, height, pitch);
}

size_t shim_audio_sample_batch(const int16_t *data, size_t frames)
{
   if (!audio_batch_cb)
      return 0;
   return audio_batch_cb(data, frames);
}

void shim_input_poll(void)
{
   if (input_poll_cb)
      input_poll_cb();
}

int16_t shim_input_state(unsigned port, unsigned device, unsigned index, unsigned id)
{
   if (!input_state_cb)
      return 0;
   return input_state_cb(port, device, index, id);
}

void shim_keyboard_callback(bool down, unsigned keycode, uint32_t character, uint16_t key_modifiers)
{
   /* izapple2KeyboardEvent is exported by keyboard.go */
   izapple2KeyboardEvent(down ? 1 : 0, keycode, character, key_modifiers);
}

/*
The disk control interface. The frontend needs C function pointers, so these
forward to the Go functions exported by disk.go.
*/

static bool disk_set_eject_state(bool ejected)
{
   return izapple2DiskSetEjectState(ejected ? 1 : 0) != 0;
}

static bool disk_get_eject_state(void)
{
   return izapple2DiskGetEjectState() != 0;
}

static unsigned disk_get_image_index(void)
{
   return izapple2DiskGetImageIndex();
}

static bool disk_set_image_index(unsigned index)
{
   return izapple2DiskSetImageIndex(index) != 0;
}

static unsigned disk_get_num_images(void)
{
   return izapple2DiskGetNumImages();
}

static bool disk_replace_image_index(unsigned index, const struct retro_game_info *info)
{
   return izapple2DiskReplaceImageIndex(index, (struct retro_game_info *)info) != 0;
}

static bool disk_add_image_index(void)
{
   return izapple2DiskAddImageIndex() != 0;
}

static bool disk_set_initial_image(unsigned index, const char *path)
{
   return izapple2DiskSetInitialImage(index, (char *)path) != 0;
}

static bool disk_get_image_path(unsigned index, char *s, size_t len)
{
   return izapple2DiskGetImagePath(index, s, len) != 0;
}

static bool disk_get_image_label(unsigned index, char *s, size_t len)
{
   return izapple2DiskGetImageLabel(index, s, len) != 0;
}

bool shim_set_disk_control_interface(void)
{
   static const struct retro_disk_control_ext_callback ext = {
      disk_set_eject_state,
      disk_get_eject_state,
      disk_get_image_index,
      disk_set_image_index,
      disk_get_num_images,
      disk_replace_image_index,
      disk_add_image_index,
      disk_set_initial_image,
      disk_get_image_path,
      disk_get_image_label
   };

   if (shim_environment(RETRO_ENVIRONMENT_SET_DISK_CONTROL_EXT_INTERFACE, (void *)&ext))
      return true;

   /* An older frontend, that knows only the first seven and has no labels */
   static const struct retro_disk_control_callback basic = {
      disk_set_eject_state,
      disk_get_eject_state,
      disk_get_image_index,
      disk_set_image_index,
      disk_get_num_images,
      disk_replace_image_index,
      disk_add_image_index
   };

   return shim_environment(RETRO_ENVIRONMENT_SET_DISK_CONTROL_INTERFACE, (void *)&basic);
}

/*
The libretro entry points that take a const pointer are defined here. cgo can
not put const in the declarations it generates for the exported Go functions,
so those four would clash with the ones in libretro.h. The rest of the
retro_ functions are exported straight from core.go.
*/

bool retro_load_game(const struct retro_game_info *game)
{
   return izapple2LoadGame((struct retro_game_info *)game);
}

bool retro_load_game_special(unsigned game_type, const struct retro_game_info *info, size_t num_info)
{
   return izapple2LoadGameSpecial(game_type, (struct retro_game_info *)info, num_info);
}

bool retro_unserialize(const void *data, size_t size)
{
   return izapple2Unserialize((void *)data, size);
}

void retro_cheat_set(unsigned index, bool enabled, const char *code)
{
   izapple2CheatSet(index, enabled, (char *)code);
}
