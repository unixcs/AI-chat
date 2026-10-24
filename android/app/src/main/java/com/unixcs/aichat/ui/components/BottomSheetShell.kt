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
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.input.pointer.util.VelocityTracker
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import kotlinx.coroutines.launch

/**
 * Hand-rolled bottom sheet — same rationale as [DrawerShell]: material3's
 * ModalBottomSheet rides the same AnchoredDraggable machinery that froze
 * drawers on real devices.
 *
 * Lifecycle guarantee: every close path (scrim tap, back, drag release) runs
 * the watchdogged [Animatable.settleTo] and, once landed at 0, fires
 * [onClosed] exactly once — so a caller keeping the sheet mounted behind a
 * boolean (e.g. rememberSaveable) is always told to unmount it. Without that
 * callback the sheet used to stay mounted-but-invisible and its open button
 * became a no-op forever. Drag-to-dismiss lives only on the grab handle, so
 * it never fights the scrollable body.
 *
 * Recomposition discipline: per-frame values are read ONLY inside
 * graphicsLayer lambdas (draw phase); the composition-visible booleans go
 * through derivedStateOf and flip once per transition, not per frame.
 *
 * @param progress caller-owned 0..1 animation; 0 = hidden, 1 = shown.
 * @param onClosed invoked after a gesture-driven settle lands at hidden.
 * @param sheet the sheet body.
 * @param content the main surface underneath; empty when the caller only
 *     overlays the sheet on top of already-composed siblings.
 */
@Composable
fun BottomSheetShell(
    progress: Animatable<Float, *>,
    onClosed: () -> Unit = {},
    sheet: @Composable () -> Unit,
    content: @Composable () -> Unit = {},
) {
    val scope = rememberCoroutineScope()
    val density = LocalDensity.current
    val flickPx = with(density) { FLICK_VELOCITY.dp.toPx() }
    // Real window height from this node's layout (includes system bars under
    // edge-to-edge, which LocalConfiguration.screenHeightDp excludes), read at
    // event time so drags stay correct across resizes without restarting the
    // gesture detector mid-stream.
    val heightPxState = remember { mutableFloatStateOf(0f) }
    val tracker = remember { VelocityTracker() }
    // Guards the exactly-once contract of [onClosed]: two close paths firing
    // in the same gesture window (scrim tap + back) must not report twice.
    var closedReported by remember { mutableStateOf(false) }

    val open by remember { derivedStateOf { progress.value > 0f } }

    fun close(velocityY: Float = 0f) {
        scope.launch {
            // Downward flick dismisses regardless of position; upward flick
            // re-shows; otherwise the position threshold decides.
            val target = when {
                velocityY > flickPx -> 0f
                velocityY < -flickPx -> 1f
                else -> if (progress.value > SETTLE_THRESHOLD) 1f else 0f
            }
            progress.settleTo(target)
            if (target == 1f) {
                closedReported = false
            } else if (!closedReported) {
                closedReported = true
                onClosed()
            }
        }
    }

    // Reads progress.value INSIDE the coroutine: per-frame pointer batching
    // makes two same-frame absolute snaps overwrite each other's delta.
    fun dragTo(deltaPx: Float) {
        scope.launch {
            if (heightPxState.floatValue > 0f) {
                progress.snapTo((progress.value - deltaPx / heightPxState.floatValue).coerceIn(0f, 1f))
            }
        }
    }

    BackHandler(enabled = open) { close() }

    Box(
        Modifier
            .fillMaxSize()
            .onSizeChanged { heightPxState.floatValue = it.height.toFloat() },
    ) {
        content()

        if (open) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .zIndex(4f)
                    .graphicsLayer { alpha = progress.value }
                    .background(Color.Black.copy(alpha = 0.45f))
                    .clearAndSetSemantics {}
                    .pointerInput(Unit) {
                        detectTapGestures(onTap = { close() })
                    },
            )

            Surface(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .zIndex(5f)
                    .graphicsLayer { translationY = heightPxState.floatValue * (1 - progress.value) },
                shape = RoundedCornerShape(topStart = 24.dp, topEnd = 24.dp),
                color = MaterialTheme.colorScheme.surface,
            ) {
                Column(Modifier.fillMaxWidth()) {
                    // Grab handle: the only drag-to-dismiss surface, so the
                    // scrollable sheet body keeps gestures untouched.
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .pointerInput(Unit) {
                                detectVerticalDragGestures(
                                    onDragStart = { tracker.resetTracking() },
                                    onDragEnd = {
                                        close(tracker.calculateVelocity().y)
                                        tracker.resetTracking()
                                    },
                                    onDragCancel = {
                                        close()
                                        tracker.resetTracking()
                                    },
                                ) { change, dragAmount ->
                                    tracker.addPosition(change.uptimeMillis, change.position)
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
