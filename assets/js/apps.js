if ($('#apps-page').length > 0) {

    const $chosens = $('select.form-control-chosen');
    const $table = $('table.table-datatable2');
    const $form = $('form');

    // Set form fields from URL
    if (window.location.hash) {
        $form.deserialize(window.location.hash.substr(1));
    }

    // Setup drop downs
    $chosens.chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false
    });

    // Sliders
    const priceElement = document.getElementById('price-slider');
    const priceHigh = parseInt($(priceElement).attr('data-high'));
    const priceSlider = noUiSlider.create(document.getElementById('price-slider'), {
        start: [0, priceHigh],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': priceHigh
        }
    });

    const scoreElement = document.getElementById('score-slider');
    const scoreSlider = noUiSlider.create(document.getElementById('score-slider'), {
        start: [0, 100],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100
        }
    });

    // Form changes
    $chosens.on('change', filter);
    $form.on('submit', filter);
    priceSlider.on('set.one', filter);
    scoreSlider.on('set.one', filter);

    function filter(e) {
        $table.DataTable().draw();
        history.pushState({}, document.title, "/games#" + $form.serialize().replace('name=&', ''));
        updateLabels(e);
        return false;
    }

    // Slider labels
    $(document).ready(updateLabels);

    function updateLabels(e) {

        const prices = priceSlider.get();
        const scores = scoreSlider.get();

        $('label#price-label').html('Price ($' + Math.round(prices[0]) + ' - $' + Math.round(prices[1]) + ')');
        $('label#score-label').html('Score (' + Math.round(scores[0]) + '% - ' + Math.round(scores[1]) + '%)');
    }

    // Setup datatable
    $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            delete data.columns;
            delete data.length;
            delete data.search.regex;

            data.search.tags = $('#tags').val();
            data.search.genres = $('#genres').val();
            data.search.developers = $('#developers').val();
            data.search.publishers = $('#publishers').val();
            data.search.platforms = $('#platforms').val();
            data.search.types = $('#types').val();
            data.search.search = $('#search').val();
            data.search.prices = priceSlider.get();
            data.search.scores = scoreSlider.get();

            $.ajax({
                url: $(this).attr('data-path'),
                data: data,
                success: callback,
                dataType: 'json',
                cache: true
            });
        },
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                }
            },
            // Type
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false
            },
            // Score
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5] + '%';
                }
            },
            // DLC Count
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6];
                }
            },
            // Price
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '$' + row[7];
                }
            }
        ]
    }));
}
