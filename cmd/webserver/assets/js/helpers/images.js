function observeLazyImages($target) {

    // https://www.sitepoint.com/five-techniques-lazy-load-images-website-performance/

    if (!$target) {
        return;
    }

    if (typeof $target === 'string' || typeof $target === 'object') {
        $target = $($target);
    }

    const callback = function (entries, self) {

        // iterate over each entry
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                loadImage($(entry.target));
                self.unobserve(entry.target);
            }
        });
    };

    const config = {
        rootMargin: '0px 0px 50px 0px',
        threshold: 0
    };

    let observer = new IntersectionObserver(callback, config);

    $target.each(function (index) {
        observer.observe(this);
    });
}

observeLazyImages('img[data-lazy]');

function loadImage($target) {

    const $alt = $target.attr('data-lazy-alt');
    if ($alt) {
        $target.attr('alt', $alt)
    }

    const $title = $target.attr('data-lazy-title');
    if ($title) {
        $target.attr('title', $title)
    }

    $target.attr('src', $target.attr('data-lazy'));

    //
    $target.removeAttr("data-lazy-alt");
    $target.removeAttr("data-lazy-title");
    $target.removeAttr("data-lazy");
}

function fixBrokenImages() {

    // This can't be on document as img events dont bubble up.
    $('img').one('error', function () {

        const url = $(this).attr('data-src');
        if (url) {
            this.src = url;
        } else {
            this.src = '/assets/img/no-app-image-square.jpg';
        }
    });

    $('img[src=""][data-src]').each(function (i, value) {
        this.src = $(this).attr('data-src');
    });
}

$(document).ready(fixBrokenImages);
