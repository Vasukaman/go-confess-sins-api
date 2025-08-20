// in /static/gachaAnimator.js

export class GachaAnimator {
    
    // 2. Animates the roll button press
    animateRollButtonPress(buttonElement, onComplete) {
        anime({
            targets: buttonElement,
            translateY: '25%', // Move it down by its full height
            duration: 200,
            easing: 'easeInQuad',
            complete: onComplete // This function will be called when the animation finishes
        });
    }

    // Animates the roll button return
    animateRollButtonReturn(buttonElement) {
        anime({
            targets: buttonElement,
            translateY: '-30%', // Move it back to its original position
            duration: 300,
            easing: 'easeOutQuad'
        });
    }

    // 4. Shakes the entire gacha container
    animateGachaShake(containerElement) {
        anime({
            targets: containerElement,
            duration: 500,
            easing: 'easeInOutSine',
            translateX: [
                { value: -2, duration: 50 },
                { value: 2, duration: 50 },
                { value: -1, duration: 50 },
                { value: 1, duration: 50 },
                { value: 0, duration: 50 }
            ],
            translateY: [ // Shake more vertically
                { value: 4, duration: 50 },
                { value: -4, duration: 50 },
                { value: 3, duration: 50 },
                { value: -3, duration: 50 },
                { value: 0, duration: 50 }
            ],
            loop: 2 // Loop the shake twice
        });
    }

    // 3. Animates the display circle in and out
    animateDisplayCircle(containerElement, shouldShow) {
        anime({
            targets: containerElement,
            translateX: shouldShow ? '0%' : '-30%', // Move in from right or out to left
            duration: 500,
            easing: 'easeOutExpo'
        });
    }
}