;(function ($, window, document, user, undefined) {

    "use strict";

    // Create the defaults once
    const pluginName = "gdbTable";
    const defaults = {
        cache: true,
        searchFields: [],
        tableOptions: {
            "autoWidth": false,
            "dom": '<"dt-pagination"p>t<"dt-pagination"p>r',
            "fixedHeader": true,
            "info": false,
            "processing": false,
            "language": {
                "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
                "paginate": {
                    "next": '<i class="fas fa-chevron-right"></i>',
                    "previous": '<i class="fas fa-chevron-left"></i>',
                },
            },
            "lengthChange": false,
            "ordering": true,
            "pageLength": 100,
            "paging": true,
            "pagingType": 'simple_numbers',
            "searching": true,
            "stateSave": false,
        },
    };

    // The actual plugin constructor
    function Plugin(element, options) {

        if (options == null) {
            options = {}
        }

        if (options.tableOptions == null) {
            options.tableOptions = {};
        }

        options.isAjax = function () {
            return $(element).attr('data-path') != null;
        };

        if (options.isAjax()) {

            defaults.tableOptions.processing = false;
            defaults.tableOptions.serverSide = true;
            defaults.tableOptions.orderMulti = false;

            const parentSettings = $.extend(true, {}, defaults, options);

            defaults.tableOptions.ajax = function (data, callback, settings) {

                delete data.columns;
                data.search = {};

                // Add search fields to ajax query from URL
                const params = new URL(window.location).searchParams;
                for (const $field of parentSettings.searchFields) {

                    let name, value = '';

                    if ($field.prop('multiple')) {

                        // Multi select
                        name = $field.attr('name');
                        value = params.getAll(name);

                    } else if ($field.hasClass('noUi-target')) {

                        // Slider
                        name = $field.attr('data-name');
                        value = params.getAll(name);

                    } else { // Inputs

                        name = $field.attr('name');
                        value = params.get(name);

                    }

                    if (name && value && value.length > 0) {
                        data.search[name] = value;
                    }
                }

                $.ajax({
                    data: data,
                    dataType: 'json',
                    cache: options.cache,
                    url: function () {
                        return $(element).attr('data-path');
                    }(),
                    success: callback,
                    error: function (jqXHR, textStatus, errorThrown) {

                        data = {
                            "draw": "1",
                            "recordsTotal": "0",
                            "recordsFiltered": "0",
                            "data": [],
                            "limited": false
                        };

                        callback(data, textStatus, null);
                    },
                });
            }
        } else {

            defaults.tableOptions.search = {
                "smart": true
            };

            defaults.tableOptions.columnDefs = [
                {
                    "orderable": false,
                    "targets": $(element).find('thead tr th[data-disabled]').map(function () {
                        return $(this).index();
                    }).get(),
                }
            ]
        }

        //
        this.settingsWithoutUrl = $.extend(true, {}, defaults, options);

        // Add url params to options
        const urlOptions = {};
        const params = new URL(window.location).searchParams;
        if (params.get('page')) {
            urlOptions.displayStart = (params.get('page') - 1) * this.settingsWithoutUrl.tableOptions.pageLength;
        }
        if (params.get('sort') && params.get('order')) {
            urlOptions.order = [[parseInt(params.get('sort')), params.get('order')]];
        }

        //
        this.settings = $.extend(true, {}, defaults, options, {tableOptions: urlOptions});
        this.element = element;
        this.user = user;
        this.init();
    }

    $.extend(Plugin.prototype, {
        init: function () {

            const parent = this;

            // Before AJAX
            $(this.element).on('preXhr.dt', function (e, settings, data) {

                // Fade
                $(parent.element).fadeTo(500, 0.3);
            });

            // After AJAX
            $(this.element).on('xhr.dt', function (e, settings, json, xhr) {

                // Fade
                $(parent.element).fadeTo(100, 1);

                // Add donate button
                parent.limited = json.limited;
            });

            // Init table
            // console.log(parent.element, this.settings.tableOptions);
            const dt = $(this.element).DataTable(this.settings.tableOptions);
            this.dt = dt; // To return from plugin call

            // On Draw
            $(this.element).on('draw.dt', function (e, settings) {

                // Add donate button
                if (parent.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">See more!</span></a></small></li>');
                    $(parent.element).parent().find('.dt-pagination ul.pagination').append(donate);
                }

                // Hide empty pagination
                const $pagination = $(parent.element).parent().find('.dt-pagination');
                (dt.page.info().pages <= 1)
                    ? $pagination.hide()
                    : $pagination.show();

                // Update URL
                if ($(parent.element).is(":visible")) {

                    const order = dt.order();
                    const settingsOrder = parent.settingsWithoutUrl.tableOptions.order;

                    if (settingsOrder != null && order.length > 0) {
                        if (order[0][1] === settingsOrder[0][1] && order[0][0] === settingsOrder[0][0]) {
                            deleteUrlParam('order');
                            deleteUrlParam('sort');
                        } else {
                            setUrlParam('order', order[0][1]);
                            setUrlParam('sort', order[0][0]);
                        }
                    }

                    if (dt.page.info().page === 0) {
                        deleteUrlParam('page');
                    } else {
                        setUrlParam('page', dt.page.info().page + 1);
                    }
                }

                // Bold rows
                parent.highlightRows();

                // Lazy load images
                observeLazyImages($(parent.element).find('img[data-lazy]'));

                // Fix broken images
                fixBrokenImages();
            });

            // Hydrate search field inputs from url params
            const params = new URL(window.location).searchParams;
            for (const $field of this.settings.searchFields) {

                if ($field.hasClass('noUi-target')) { // Slider

                    const slider = $field[0].noUiSlider;
                    const name = $field.attr('data-name');

                    if (params.has(name)) {
                        slider.set(params.getAll(name));
                    }

                } else { // Input

                    const name = $field.attr('name');

                    if (params.has(name)) {

                        $field.val(params.getAll(name));

                        // Update Chosen drop downs
                        if ($field.hasClass('form-control-chosen')) {
                            $field.trigger("chosen:updated");
                        }
                    }

                }
            }

            // On page change
            dt.on('page.dt', function (e, settings, processing) {

                // Scroll on pagination click
                let padding = 15;
                if ($('.fixedHeader-floating').length > 0) {
                    padding = padding + 48;
                }
                $('html, body').animate({
                    scrollTop: $(this).prev().offset().top - padding
                }, 200);
            });

            // On tab change
            $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

                // Fixes hidden fixed header tables
                $.each(window.gdbTables, function (index, value) {
                    value.fixedHeader.adjust();
                });

                //
                clearUrlParams();
            });

            // Update URL when search fields are changed
            if (this.settings.isAjax()) {
                for (const $field of this.settings.searchFields) {

                    if ($field.hasClass('noUi-target')) { // Slider

                        const slider = $field[0].noUiSlider;
                        const name = $field.attr('data-name');

                        slider.on('set', function (e) {

                            const value = slider.get();

                            if (name && value) {
                                setUrlParam(name, value);
                            } else {
                                deleteUrlParam(name);
                            }

                            if (typeof updateLabels == 'function') {
                                updateLabels();
                            }

                            dt.draw();
                        });


                        slider.on('update', function (e) {
                            if (typeof updateLabels == 'function') {
                                updateLabels();
                            }
                        });

                    } else { // Inputs

                        const name = $field.attr('name');

                        $field.on('change search', function (e) {

                            const value = $field.val();

                            if (name && value) {
                                setUrlParam(name, value);
                            } else {
                                deleteUrlParam(name);
                            }

                            dt.draw();

                            return false;
                        });
                    }
                }
            } else {
                for (const $field of this.settings.searchFields) {
                    $field.on('change search', function (e) {
                        dt.search($(this).val());
                        dt.draw();
                    });
                }
            }

            // Fixes scrolling to pagination on every click
            $(this.element).parent().find(".paginate_button > a").one("focus", function () {
                $(this).blur();
            });

            // Local tables finish initializing before event handlers are attached,
            // so we trigger them again here.
            if (!this.settings.isAjax()) {
                $(parent.element).trigger('draw.dt');
            }

            // Keep track of tables, so we can recalculate fixed headers on tab changes etc
            window.gdbTables = window.gdbTables || [];
            window.gdbTables.push();
        },
        highlightRows: function () {

            if (this.user.isLoggedIn) {
                let games = localStorage.getItem('games');
                if (games != null) {
                    games = JSON.parse(games);
                    if (games != null) {
                        $('[data-app-id]').each(function () {
                            const id = $(this).attr('data-app-id');
                            if (games.includes(parseInt(id))) {
                                $(this).addClass('font-weight-bold')
                            }
                        });
                    }
                }

                let groups = localStorage.getItem('groups');
                if (groups != null) {
                    groups = JSON.parse(groups);
                    if (groups != null) {
                        $('[data-group-id]').each(function () {
                            const id = $(this).attr('data-group-id');
                            if (groups.includes(id)) {
                                $(this).addClass('font-weight-bold')
                            }
                            const id64 = $(this).attr('data-group-id64');
                            if (groups.includes(id64)) {
                                $(this).addClass('font-weight-bold')
                            }
                        });
                    }
                }
            }
        },
    });

    $.fn[pluginName] = function (options) {
        return new Plugin(this, options).dt;
    };

    // Init local tables
    $('table.table.table-datatable').each(function (index) {
        $(this).gdbTable();
    });


})(jQuery, window, document, user);
