/* global gsap, ScrollTrigger */
(function () {
  'use strict';

  if (typeof gsap === 'undefined') {
    return;
  }

  var reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  if (reduced) {
    document.documentElement.classList.add('motion-reduced');
    return;
  }

  gsap.registerPlugin(ScrollTrigger);

  var easeOut = 'power3.out';
  var easeInOut = 'power2.inOut';

  /* ── Nav ── */
  gsap.from('.nav-inner', {
    y: -16,
    opacity: 0,
    duration: 0.6,
    ease: easeOut,
  });

  /* ── Hero entrance ── */
  var heroTl = gsap.timeline({ defaults: { ease: easeOut } });

  heroTl
    .from('.hero-badge', { y: 20, opacity: 0, duration: 0.5 })
    .from('.hero-headline', { y: 28, opacity: 0, duration: 0.65 }, '-=0.25')
    .from('.hero-sub', { y: 20, opacity: 0, duration: 0.5 }, '-=0.35')
    .from('.install-block', { y: 24, opacity: 0, duration: 0.55 }, '-=0.3')
    .from('.hero-cta', { y: 12, opacity: 0, duration: 0.4 }, '-=0.2')
    .from(
      '.hero-terminal',
      { y: 40, opacity: 0, scale: 0.97, duration: 0.8, ease: easeInOut },
      '-=0.55'
    );

  /* Terminal typewriter-style line reveal */
  var lines = gsap.utils.toArray('.terminal-body .tl:not(.spacer)');
  if (lines.length) {
    heroTl.to(
      lines,
      {
        opacity: 1,
        duration: 0.12,
        stagger: 0.07,
        ease: 'none',
      },
      '-=0.35'
    );
  }

  /* Cursor blink starts after terminal finishes */
  heroTl.add(function () {
    var cursor = document.querySelector('.terminal-body .cursor');
    if (cursor) {
      cursor.classList.add('cursor-active');
    }
  });

  /* ── Scroll reveals ── */
  function revealOnScroll(selector, options) {
    var targets = gsap.utils.toArray(selector);
    if (!targets.length) {
      return;
    }
    gsap.from(targets, {
      scrollTrigger: {
        trigger: options.trigger || targets[0],
        start: options.start || 'top 85%',
        once: true,
      },
      y: options.y || 32,
      opacity: 0,
      duration: options.duration || 0.65,
      ease: easeOut,
      stagger: options.stagger || 0,
    });
  }

  revealOnScroll('.showcase .section-label, .showcase .section-title', {
    trigger: '.showcase',
    stagger: 0.12,
    y: 24,
  });

  revealOnScroll('.cards-grid .card', {
    trigger: '.cards-grid',
    stagger: 0.1,
    y: 36,
  });

  revealOnScroll('.features .feature', {
    trigger: '.features-grid',
    stagger: 0.08,
    y: 28,
  });

  revealOnScroll('.cmd-ref .section-label, .cmd-ref .section-title', {
    trigger: '.cmd-ref',
    stagger: 0.12,
    y: 24,
  });

  revealOnScroll('.cmd-group', {
    trigger: '.cmd-groups',
    stagger: 0.06,
    y: 20,
    duration: 0.5,
  });

  revealOnScroll('.install-full .section-label, .install-full .section-title', {
    trigger: '.install-full',
    stagger: 0.12,
    y: 24,
  });

  revealOnScroll('.install-card', {
    trigger: '.install-grid',
    stagger: 0.1,
    y: 32,
  });

  revealOnScroll('.install-note', {
    trigger: '.install-note',
    y: 16,
    duration: 0.5,
  });

  revealOnScroll('.footer-inner', {
    trigger: '.footer',
    y: 20,
    duration: 0.55,
  });

  /* Subtle parallax on hero terminal while in view */
  gsap.to('.hero-terminal', {
    scrollTrigger: {
      trigger: '.hero',
      start: 'top top',
      end: 'bottom top',
      scrub: 0.6,
    },
    y: 48,
    ease: 'none',
  });

})();