/*
Bridge between the libretro C API and Go.

Go can not call the function pointers that the libretro frontend hands to the
core, so every callback is stored here and reached through a plain C function.
The keyboard callback goes the other way: the frontend needs a C function
pointer, so shim_keyboard_callback forwards to the Go side.
*/

#ifndef IZAPPLE2_SHIM_H
#define IZAPPLE2_SHIM_H

#include "libretro.h"

/* Store the callbacks given by the frontend */
void shim_set_environment(retro_environment_t cb);
void shim_set_video_refresh(retro_video_refresh_t cb);
void shim_set_audio_sample_batch(retro_audio_sample_batch_t cb);
void shim_set_input_poll(retro_input_poll_t cb);
void shim_set_input_state(retro_input_state_t cb);

/* Call the callbacks given by the frontend */
bool shim_environment(unsigned cmd, void *data);
void shim_video_refresh(const void *data, unsigned width, unsigned height, size_t pitch);
size_t shim_audio_sample_batch(const int16_t *data, size_t frames);
void shim_input_poll(void);
int16_t shim_input_state(unsigned port, unsigned device, unsigned index, unsigned id);

/* The keyboard callback the core registers with the frontend */
void shim_keyboard_callback(bool down, unsigned keycode, uint32_t character, uint16_t key_modifiers);

/* Offer the disk control interface, extended if the frontend knows it */
bool shim_set_disk_control_interface(void);

/*
The log of the frontend. It is a variadic function, which Go can not call, so
the message is formatted on the Go side and passed as a single string.
*/
void shim_set_log(retro_log_printf_t cb);
bool shim_has_log(void);
void shim_log(int level, const char *message);

#endif
