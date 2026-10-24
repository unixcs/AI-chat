package com.unixcs.aichat.ui.components

import androidx.activity.compose.BackHandler
import androidx.compose.animation.core.Animatable
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.gestures.detectVerticalDragGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import kotlinx.coroutines.launch

/**
 * Hand-rolled bottom sheet — same rationale as [DrawerShell]: material3's
 * ModalBottomSheet rides the same AnchoredDraggable machinery that froze
 * drawers on real devices. Drag-to-dismiss lives only on the grab handle, so
 * it never fights the scrollable body (a nestedScroll handshake would be the
 * alternative, at higher complexity for no UX gain here). Release always
 * settles shown or hidden — never stranded half-way.
 *
 * @param progress caller-owned 0..1 animation; 0 = hidden, 1 = shown.
 * @param onDismiss called when a gesture asks to close; the caller animates.
 * @param sheet the sheet body.
 * @param content the main surface underneath; empty when the caller only
 *     overlays the sheet on top of already-composed siblings.
 */
@Composable
fun BottomSheetShell(
    progress: Animatable<Float, *>,
    onDismiss: () -> Unit,
    sheet: @Composable () -> Unit,
    content: @Composable () -> Unit = {},
) {
    val scope = rememberCoroutineScope()
    // Screen height in px from real display metrics: the shell overlays the
    // caller's tree, so there is no own layout node to read a size from
    // before first frame (and translationY needs a value immediately).
    val configuration = androidx.compose.ui.platform.LocalConfiguration.current
    val screenHeightPx = with(LocalDensity.current) {
        configuration.screenHeightDp.dp.toPx()
    }

    fun settle() {
        scope.launch {
            progress.animateTo(
                if (progress.value > SETTLE_THRESHOLD) 1f else 0f,
                ShellSpring,
            )
        }
    }

    // Reads progress.value INSIDE the coroutine for the same reason as
    // [DrawerShell]: per-frame pointer batching makes stale absolute snaps
    // drop deltas on fast drags.
    fun dragTo(deltaPx: Float) {
        scope.launch {
            progress.snapTo((progress.value - deltaPx / screenHeightPx).coerceIn(0f, 1f))
        }
    }

    BackHandler(enabled = progress.value > 0f) { onDismiss() }

    Box(Modifier.fillMaxSize()) {
        content()

        if (progress.value > 0f) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .zIndex(4f)
                    .graphicsLayer { alpha = progress.value }
                    .background(Color.Black.copy(alpha = 0.45f))
                    .clearAndSetSemantics {}
                    .pointerInput(Unit) {
                        detectTapGestures(onTap = { onDismiss() })
                    },
            )

            Surface(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .zIndex(5f)
                    .graphicsLayer { translationY = screenHeightPx * (1 - progress.value) },
                shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp),
                color = MaterialTheme.colorScheme.surface,
            ) {
                Column(Modifier.fillMaxWidth()) {
                    // Grab handle: the only drag-to-dismiss surface, so the
                    // scrollable sheet body keeps gestures untouched.
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .pointerInput(screenHeightPx) {
                                detectVerticalDragGestures(
                                    onDragEnd = { settle() },
                                    onDragCancel = { settle() },
                                ) { change, dragAmount ->
                                    change.consume()
                                    dragTo(dragAmount)
                                }
                            },
                        contentAlignment = Alignment.Center,
                    ) {
                        Box(
                            Modifier
                                .padding(top = 10.dp, bottom = 6.dp)
                                .width(36.dp)
                                .height(4.dp)
                                .background(
                                    MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f),
                                    RoundedCornerShape(2.dp),
                                ),
                        )
                    }
                    Box(Modifier.navigationBarsPadding()) {
                        sheet()
                    }
                }
            }
        }
    }
}
