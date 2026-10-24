package com.unixcs.aichat.ui.components

import androidx.activity.compose.BackHandler
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.SpringSpec
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import kotlinx.coroutines.launch

/**
 * Settle/open/close spring shared by the hand-rolled shells: quick, no
 * bounce. Callers animating progress from outside must use the same spec so
 * drags, back-press and buttons all feel like one mechanism.
 */
val ShellSpring: SpringSpec<Float> = spring(
    dampingRatio = Spring.DampingRatioNoBouncy,
    stiffness = Spring.StiffnessMediumLow,
)

/** Drag past this fraction settles open; below it settles closed. */
internal const val SETTLE_THRESHOLD = 0.5f

/**
 * Hand-rolled modal drawer. Deliberately replaces material3's
 * ModalNavigationDrawer: that component rides AnchoredDraggableState, whose
 * offset is NaN until anchors are set, and it also hooks predictive-back —
 * on real devices (m3 1.3.0 and 1.3.1 both) it produced blank frozen screens
 * escapable only by killing the app, most reproducibly right after a theme
 * switch (enableEdgeToEdge re-asserts window flags in the same frame window).
 * A single Animatable progress (0 = closed, 1 = open) has no state machine to
 * wedge, and back handling is a plain BackHandler we own.
 *
 * Must wrap the whole chat Scaffold (including its bottom bar) so the scrim
 * covers the input row. The main content stays put underneath — DeepSeek
 * style: only the scrim fades and the sheet slides in.
 *
 * Drag lifecycle care: the edge hotspot would normally leave composition the
 * moment progress leaves 0, cancelling the in-flight gesture; an
 * [SETTLE_THRESHOLD]-style dragging flag keeps it mounted until release, and
 * release always settles open or closed — never stranded half-way.
 *
 * Accessibility: while the drawer is open the covered content drops its
 * semantics, and while closed the off-screen panel drops its own, so TalkBack
 * can only ever focus what is actually visible.
 *
 * @param progress caller-owned 0..1 animation; 0 = closed, 1 = fully open.
 * @param onDismiss called when a gesture asks to close; the caller animates.
 * @param sheet the drawer panel body.
 * @param content the main surface underneath.
 */
@Composable
fun DrawerShell(
    progress: Animatable<Float, *>,
    onDismiss: () -> Unit,
    sheet: @Composable () -> Unit,
    content: @Composable () -> Unit,
) {
    val scope = rememberCoroutineScope()
    val density = LocalDensity.current
    val widthPxState = remember { androidx.compose.runtime.mutableFloatStateOf(0f) }
    // Keeps the edge hotspot mounted while its drag is running: unmounting
    // mid-gesture would cancel the pointer stream and strand the drawer.
    var edgeDragging by remember { mutableStateOf(false) }

    fun settle() {
        scope.launch {
            progress.animateTo(
                if (progress.value > SETTLE_THRESHOLD) 1f else 0f,
                ShellSpring,
            )
        }
    }

    // Reads progress.value INSIDE the coroutine: pointer events are batched
    // per frame, so two onDrag calls in one frame would otherwise both compute
    // from the same stale value and the second absolute snapTo would drop the
    // first's delta (visible rubber-banding on fast drags).
    fun dragTo(deltaPx: Float, widthPx: Float) {
        scope.launch {
            if (widthPx > 0f) {
                progress.snapTo((progress.value + deltaPx / widthPx).coerceIn(0f, 1f))
            }
        }
    }

    BackHandler(enabled = progress.value > 0f) { onDismiss() }

    Box(
        Modifier
            .fillMaxSize()
            .onSizeChanged { size ->
                // Panel width in px: the narrower of 82% of the container and
                // 340dp, matching the old ModalDrawerSheet cap.
                val capPx = with(density) { 340.dp.toPx() }
                widthPxState.floatValue = (size.width * 0.82f).coerceAtMost(capPx)
            },
    ) {
        // Main surface; semantics drop while the drawer covers it.
        Box(
            Modifier
                .fillMaxSize()
                .then(
                    if (progress.value > 0.5f) {
                        Modifier.clearAndSetSemantics {}
                    } else {
                        Modifier
                    },
                ),
        ) {
            content()
        }

        // Left-edge hotspot: drag right to reveal. Vertical gestures win by
        // directional slop, so chat list scrolling stays untouched.
        if ((progress.value == 0f || edgeDragging) && widthPxState.floatValue > 0f) {
            Box(
                modifier = Modifier
                    .zIndex(1f)
                    .fillMaxHeight()
                    .width(24.dp)
                    // Keyed Unit on purpose: the handlers read widthPxState (a
                    // snapshot state) at event time, and a detector restart on
                    // resize would strand an in-flight drag without settle.
                    .pointerInput(Unit) {
                        detectHorizontalDragGestures(
                            onDragStart = { edgeDragging = true },
                            onDragEnd = {
                                edgeDragging = false
                                settle()
                            },
                            onDragCancel = {
                                edgeDragging = false
                                settle()
                            },
                        ) { change, dragAmount ->
                            if (dragAmount > 0) {
                                change.consume()
                                dragTo(dragAmount, widthPxState.floatValue)
                            }
                        }
                    }
                    .clearAndSetSemantics {},
            )
        }

        if (progress.value > 0f) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .zIndex(2f)
                    .graphicsLayer { alpha = progress.value }
                    .background(Color.Black.copy(alpha = 0.45f))
                    .clearAndSetSemantics {}
                    .pointerInput(Unit) {
                        detectTapGestures(onTap = { onDismiss() })
                    }
                    .pointerInput(Unit) {
                        detectHorizontalDragGestures(
                            onDragEnd = { settle() },
                            onDragCancel = { settle() },
                        ) { change, dragAmount ->
                            // Only a leftward drag dismisses; a rightward drag
                            // would fight the sheet's own edge.
                            if (dragAmount < 0) {
                                change.consume()
                                dragTo(dragAmount, widthPxState.floatValue)
                            }
                        }
                    },
            )
        }

        Box(
            modifier = Modifier
                .fillMaxHeight()
                .width(with(density) { widthPxState.floatValue.toDp() })
                .zIndex(3f)
                .graphicsLayer { translationX = -widthPxState.floatValue * (1 - progress.value) },
        ) {
            // Panel surface: rounded right edge, one-sided shadow, and the
            // drag-to-close gesture on the surface itself. Child taps are not
            // consumed by this detector (it only reacts to horizontal drags).
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .clip(RoundedCornerShape(topEnd = 24.dp, bottomEnd = 24.dp))
                    .background(MaterialTheme.colorScheme.surface)
                    .edgeShadow()
                    .windowInsetsPadding(WindowInsets.statusBars)
                    .then(
                        // Off-screen panel is invisible: drop its semantics so
                        // TalkBack cannot focus hidden drawer content.
                        if (progress.value > 0.5f) Modifier else Modifier.clearAndSetSemantics {},
                    )
                    .pointerInput(Unit) {
                        detectHorizontalDragGestures(
                            onDragEnd = { settle() },
                            onDragCancel = { settle() },
                        ) { change, dragAmount ->
                            if (dragAmount < 0) {
                                change.consume()
                                dragTo(dragAmount, widthPxState.floatValue)
                            }
                        }
                    },
            ) {
                sheet()
            }
        }
    }
}

/** Gradient shadow hugging the panel's right edge (drawn inside the clip). */
private fun Modifier.edgeShadow(): Modifier = drawWithContent {
    drawContent()
    val shadowWidth = 24f
    drawRect(
        brush = Brush.horizontalGradient(
            colors = listOf(Color.Transparent, Color.Black.copy(alpha = 0.14f)),
            startX = size.width - shadowWidth,
            endX = size.width,
        ),
        topLeft = Offset(size.width - shadowWidth, 0f),
        size = Size(shadowWidth, size.height),
    )
}
