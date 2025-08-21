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

      animateSlotSelect(element) {
        anime({
            targets: element,
            rotate: [0, 10, -10, 5, -5, 0], // A little jiggle and rotation
            scale: [1, 1.1, 1],
            duration: 400,
            easing: 'easeInOutSine'
        });
    }

    animateSlotDeselect(element) {
        anime({
            targets: element,
            rotate: 0,
            scale: 1,
            duration: 300,
            easing: 'easeOutQuad'
        });
    }


      animateSwap(sourceEl, destEl, onComplete) {
        if (!sourceEl || !destEl || !sourceEl.textContent || !destEl.textContent) {
            onComplete(); // If a slot is empty, just complete immediately
            return;
        }

        const sourceRect = sourceEl.getBoundingClientRect();
        const destRect = destEl.getBoundingClientRect();

        // Create temporary "clone" elements for just the emojis
        const cloneSource = document.createElement('div');
        const cloneDest = document.createElement('div');

        // Style the clones to look like the emojis
        [cloneSource, cloneDest].forEach(clone => {
            clone.style.position = 'fixed';
            clone.style.zIndex = '1000';
            clone.style.fontSize = '36px'; // Match the slot font size
            clone.style.pointerEvents = 'none'; // Prevent interaction
        });

        cloneSource.textContent = sourceEl.textContent;
        cloneSource.style.left = `${sourceRect.left + (sourceRect.width / 2) - 18}px`; // Center the clone
        cloneSource.style.top = `${sourceRect.top + (sourceRect.height / 2) - 18}px`;

        cloneDest.textContent = destEl.textContent;
        cloneDest.style.left = `${destRect.left + (destRect.width / 2) - 18}px`;
        cloneDest.style.top = `${destRect.top + (destRect.height / 2) - 18}px`;

        document.body.appendChild(cloneSource);
        document.body.appendChild(cloneDest);

        // Hide original emojis during animation
        sourceEl.style.opacity = '0';
        destEl.style.opacity = '0';

        // Animate the clones (NO SCALING)
        anime({
            targets: cloneSource,
            left: destRect.left + (destRect.width / 2) - 18,
            top: destRect.top + (destRect.height / 2) - 18,
            duration: 400,
            easing: 'easeOutExpo'
        });

        anime({
            targets: cloneDest,
            left: sourceRect.left + (sourceRect.width / 2) - 18,
            top: sourceRect.top + (sourceRect.height / 2) - 18,
            duration: 400,
            easing: 'easeOutExpo',
            complete: () => {
                cloneSource.remove();
                cloneDest.remove();
                sourceEl.style.opacity = '1';
                destEl.style.opacity = '1';
                onComplete();
            }
        });
    }


}