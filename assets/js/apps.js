if ($('#apps-page').length > 0) {

    $('#user-currency-symbol').append(' ' + user.userCurrencySymbol);

    const $chosens = $('select.form-control-chosen');
    const $table = $('table.table-datatable2');
    const $form = $('form');

    // Set form fields from URL
    if (window.location.search) {
        $form.deserialize(window.location.search.substr(1));
    }

    // Setup drop downs
    $chosens.chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false,
        max_selected_options: 10
    });

    // Setup Sliders
    const priceLow = $('#price-low').val();
    const priceHigh = $('#price-high').val();
    const priceElement = $('#price-slider')[0];
    const priceSlider = noUiSlider.create(priceElement, {
        start: [
            parseInt(priceLow ? priceLow : 0),
            parseInt(priceHigh ? priceHigh : 100)
        ],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100,
        }
    });

    const scoreLow = $('#score-low').val();
    const scoreHigh = $('#score-high').val();
    const scoreElement = $('#score-slider')[0];
    const scoreSlider = noUiSlider.create(scoreElement, {
        start: [
            parseInt(scoreLow ? scoreLow : 0),
            parseInt(scoreHigh ? scoreHigh : 100)
        ],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100
        }
    });

    // Form changes
    $chosens.on('change', redrawTable);
    $form.on('submit', redrawTable);
    priceSlider.on('set', onPriceChange);
    priceSlider.on('update', updateLabels);
    scoreSlider.on('set', onScoreChange);
    scoreSlider.on('update', updateLabels);

    function onPriceChange(e) {
        const prices = priceSlider.get();
        $('#price-low').val(prices[0]);
        $('#price-high').val(prices[1]);
        redrawTable();
    }

    function onScoreChange(e) {
        const scores = scoreSlider.get();
        $('#score-low').val(scores[0]);
        $('#score-high').val(scores[1]);
        redrawTable();
    }

    function redrawTable(e) {

        // Filter out empty form fields
        let formData = $form.serializeArray();
        formData = $.grep(formData, function (v) {
            return v.value !== "";
        });

        $table.DataTable().draw();
        history.pushState({}, document.title, "/apps?" + $.param(formData));
        updateLabels(e);
        return false;
    }

    $(document).ready(updateLabels);

    function updateLabels(e) {

        const prices = priceSlider.get();
        const scores = scoreSlider.get();

        if (prices[0] === prices[1]) {
            $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ')');
        } else {
            $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ' - ' + user.userCurrencySymbol + Math.round(prices[1]) + ')');
        }

        if (scores[0] === scores[1]) {
            $('label#score-label').html('Score (' + Math.round(scores[0]) + '%)');
        } else {
            $('label#score-label').html('Score (' + Math.round(scores[0]) + '% - ' + Math.round(scores[1]) + '%)');
        }
    }

    // Setup datatable
    $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            data.search.tags = $('#tags').val();
            data.search.genres = $('#genres').val();
            data.search.categories = $('#categories').val();
            data.search.developers = $('#developers').val();
            data.search.publishers = $('#publishers').val();
            data.search.platforms = $('#platforms').val();
            data.search.types = $('#types').val();
            data.search.search = $('#search').val();
            data.search.prices = priceSlider.get();
            data.search.scores = scoreSlider.get();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Type
            // {
            //     "targets": 1,
            //     "render": function (data, type, row) {
            //         return row[4];
            //     },
            //     "createdCell": function (td, cellData, rowData, row, col) {
            //         $(td).addClass('d-none d-lg-table-cell');
            //     },
            //     "orderable": false,
            // },
            // Players
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Score
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5] + '%';
                },
                "orderSequence": ["desc"],
            },
            // Price
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6];
                },
                "orderSequence": ["desc"],
            },
        ]
    }));
}
