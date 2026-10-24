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
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
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
import androidx.compose.ui.input.pointer.util.VelocityTracker
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import kotlinx.coroutines.TimeoutCancellationException
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeout

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
 * A release faster than this (dp/s) settles in the flick direction regardless
 * of position — the m3/DeepSeek feel. A pure position threshold ignored
 * velocity, so a light quick swipe bounced back open. 150dp/s ≈ a light quick
 * flick on real hardware; slower releases fall back to the position rule.
 */
internal const val FLICK_VELOCITY = 150

/**
 * Release target, position- and velocity-aware: a quick flick settles in the
 * flick direction regardless of position — the m3/DeepSeek feel, and the fix
 * for the bounce-back bug (a light quick left-swipe from the open drawer must
 * close instead of springing back). Drawer semantics only; BottomSheetShell
 * maps vertical velocities itself (up = show, down = hide). Idle taps never
 * reach this: a tap does not cross drag slop, so no release handler fires.
 */
internal fun settleTarget(value: Float, velocity: Float, flickPx: Float): Float = when {
    velocity > flickPx -> 1f
    velocity < -flickPx -> 0f
    else -> if (value > SETTLE_THRESHOLD) 1f else 0f
}

/**
 * Watchdog window for one settle. If the frame clock stalls (OEM-side hiccups
 * are the recurring "blank frozen screen" class), the animation coroutine
 * would otherwise never complete and the overlay stays mounted, input-eating
 * and half-visible forever. On timeout the stuck animation is cancelled and
 * [Animatable.snapTo] — a plain state write that needs no frames — finishes
 * the transition, so an unrecoverable intermediate state cannot exist.
 */
private const val SETTLE_WATCHDOG_MS = 1_200L

/**
 * Settle [progress] to [target] with the watchdog. Cancel-safe: outer scope
 * cancellation propagates normally.
 */
suspend fun Animatable<Float, *>.settleTo(target: Float) {
    try {
        withTimeout(SETTLE_WATCHDOG_MS) {
            animateTo(target, ShellSpring)
        }
    } catch (_: TimeoutCancellationException) {
        // A newer settle may have taken the animation over between the timeout
        // and here; only finish the anchor ourselves if nothing is animating,
        // so the watchdog can never override the user's latest request.
        if (!isRunning) snapTo(target)
    }
}

/**
 * Hand-rolled modal drawer. Deliberately replaces material3's
 * ModalNavigationDrawer: that component rides AnchoredDraggableState, whose
 * offset is NaN until anchors are set — on real devices it produced blank
 * frozen screens escapable only by killing the app. A single Animatable
 * progress (0 = closed, 1 = open) has no state machine to wedge.
 *
 * Recomposition discipline: per-frame values are read ONLY inside
 * graphicsLayer lambdas (draw phase). The few boolean reads in composition go
 * through derivedStateOf so they flip once per transition, not once per
 * frame — this shell wraps the whole chat Scaffold, and per-frame recomposition
 * here used to recompose the entire chat tree (the reported jank).
 *
 * Drag lifecycle care: the edge hotspot stays mounted while its drag runs
 * ([edgeDragging]); release always settles open or closed, velocity-aware,
 * with the [settleTo] watchdog — never stranded half-way.
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
    val flickPx = with(density) { FLICK_VELOCITY.dp.toPx() }
    val widthPxState = remember { mutableFloatStateOf(0f) }
    val tracker = remember { VelocityTracker() }
    // Keeps the edge hotspot mounted while its drag is running: unmounting
    // mid-gesture would cancel the pointer stream and strand the drawer.
    var edgeDragging by remember { mutableStateOf(false) }

    // Boolean projections of progress: recompose only when they flip.
    val open by remember { derivedStateOf { progress.value > 0f } }
    val closed by remember { derivedStateOf { progress.value == 0f } }
    val covered by remember { derivedStateOf { progress.value > SETTLE_THRESHOLD } }

    fun settle(velocity: Float) {
        scope.launch { progress.settleTo(settleTarget(progress.value, velocity, flickPx)) }
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

    BackHandler(enabled = open) { onDismiss() }

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
                .then(if (covered) Modifier.clearAndSetSemantics {} else Modifier),
        ) {
            content()
        }

        // Left-edge hotspot: drag right to reveal. Vertical gestures win by
        // directional slop, so chat list scrolling stays untouched.
        if ((closed || edgeDragging) && widthPxState.floatValue > 0f) {
            // If this node is disposed mid-drag (structural recomposition,
            // activity pause), neither onDragEnd nor onDragCancel runs — clear
            // the latch so the hotspot cannot stay phantom-mounted.
            DisposableEffect(Unit) { onDispose { edgeDragging = false } }
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
                            onDragStart = {
                                tracker.resetTracking()
                                edgeDragging = true
                            },
                            onDragEnd = {
                                edgeDragging = false
                                settle(tracker.calculateVelocity().x)
                                tracker.resetTracking()
                            },
                            onDragCancel = {
                                edgeDragging = false
                                settle(0f)
                                tracker.resetTracking()
                            },
                        ) { change, dragAmount ->
                            tracker.addPosition(change.uptimeMillis, change.position)
                            if (dragAmount > 0) {
                                change.consume()
                                dragTo(dragAmount, widthPxState.floatValue)
                            }
                        }
                    }
                    .clearAndSetSemantics {},
            )
        }

        if (open) {
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
                            onDragStart = { tracker.resetTracking() },
                            onDragEnd = {
                                settle(tracker.calculateVelocity().x)
                                tracker.resetTracking()
                            },
                            onDragCancel = {
                                settle(0f)
                                tracker.resetTracking()
                            },
                        ) { change, dragAmount ->
                            tracker.addPosition(change.uptimeMillis, change.position)
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
                        if (covered) Modifier else Modifier.clearAndSetSemantics {},
                    )
                    .pointerInput(Unit) {
                        detectHorizontalDragGestures(
                            onDragStart = { tracker.resetTracking() },
                            onDragEnd = {
                                settle(tracker.calculateVelocity().x)
                                tracker.resetTracking()
                            },
                            onDragCancel = {
                                settle(0f)
                                tracker.resetTracking()
                            },
                        ) { change, dragAmount ->
                            tracker.addPosition(change.uptimeMillis, change.position)
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
